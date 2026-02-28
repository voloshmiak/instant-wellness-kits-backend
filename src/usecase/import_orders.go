package usecase

import (
	"InstantWellnessKits/src/entity"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type ImportOrdersUseCase struct {
	geocodingService GeocodingService
	orders           Orders
	taxRates         TaxRates
}

func NewImportOrdersUseCase(geocodingService GeocodingService,
	orders Orders, taxRates TaxRates) *ImportOrdersUseCase {
	return &ImportOrdersUseCase{
		geocodingService: geocodingService,
		orders:           orders,
		taxRates:         taxRates,
	}
}

func (uc *ImportOrdersUseCase) Execute(ctx context.Context, fileReader io.Reader) ([]ImportResult, error) {
	reader := csv.NewReader(fileReader)

	_, _ = reader.Read()

	numWorkers := 10
	jobs := make(chan ImportJob, 100)
	results := make(chan ImportResult, 100)

	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go uc.worker(ctx, jobs, results, &wg)
	}

	go func() {
		rateLimiter := time.NewTicker(time.Second / 20)
		defer rateLimiter.Stop()

		rowNum := 2
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				results <- ImportResult{RowNumber: rowNum, Success: false,
					Err: fmt.Errorf("csv read error: %w", err)}
				rowNum++
				continue
			}

			<-rateLimiter.C

			lon, _ := strconv.ParseFloat(record[1], 64)
			lat, _ := strconv.ParseFloat(record[2], 64)
			timestamp, _ := time.Parse(time.DateTime, record[3])
			subtotal, _ := decimal.NewFromString(record[4])

			jobs <- ImportJob{
				RowNumber: rowNum,
				Latitude:  lat,
				Longitude: lon,
				Subtotal:  subtotal,
				Timestamp: timestamp,
			}
			rowNum++
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	const batchSize = 500

	var allResults []ImportResult
	var toCreate []*entity.Order

	flush := func() error {
		if len(toCreate) == 0 {
			return nil
		}
		if err := uc.orders.CreateBatch(ctx, toCreate); err != nil {
			return fmt.Errorf("batch create failed: %w", err)
		}
		toCreate = toCreate[:0]
		return nil
	}

	for res := range results {
		allResults = append(allResults, res)
		if res.Success {
			toCreate = append(toCreate, res.Order)
			if len(toCreate) >= batchSize {
				if err := flush(); err != nil {
					return allResults, err
				}
			}
		} else {
			log.Printf("Import error at row %d: %v", res.RowNumber, res.Err)
		}
	}

	if err := flush(); err != nil {
		return allResults, err
	}

	return allResults, nil
}

type ImportJob struct {
	RowNumber int
	Latitude  float64
	Longitude float64
	Subtotal  decimal.Decimal
	Timestamp time.Time
}

type ImportResult struct {
	RowNumber int
	Success   bool
	Err       error
	Order     *entity.Order
}

func (uc *ImportOrdersUseCase) worker(ctx context.Context,
	jobs <-chan ImportJob, results chan<- ImportResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		log.Println("Processing row", job.RowNumber)
		juris, err := uc.geocodingService.GetJurisdiction(job.Latitude, job.Longitude)
		if err != nil {
			results <- ImportResult{RowNumber: job.RowNumber, Success: false, Err: err}
			continue
		}

		if juris.State != "New York" {
			results <- ImportResult{RowNumber: job.RowNumber, Success: false,
				Err: fmt.Errorf("delivery location is outside New York State (got: %s)", juris.State)}
			continue
		}

		compositeTaxRate, taxBreakdown, err := uc.taxRates.Get(ctx, juris)
		if err != nil {
			results <- ImportResult{RowNumber: job.RowNumber, Success: false, Err: err}
			continue
		}

		taxAmount := job.Subtotal.Mul(compositeTaxRate).Round(2)
		totalAmount := job.Subtotal.Add(taxAmount)

		order := entity.NewOrder(job.Latitude, job.Longitude,
			job.Subtotal, compositeTaxRate, taxAmount, totalAmount,
			taxBreakdown, juris, job.Timestamp)

		results <- ImportResult{RowNumber: job.RowNumber, Success: true, Order: order}
	}
}
