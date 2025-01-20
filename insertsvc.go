package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

func InsertPackage(db *sql.DB, pkg Clientmodel) error {
	query := `
   INSERT INTO "clientmodel".clientmodel (deviceid, groupid, modelid, groupname, modelname, createdby, updateat)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (deviceid)
DO UPDATE SET
    groupid = EXCLUDED.groupid,
    modelid = EXCLUDED.modelid,
    groupname = EXCLUDED.groupname,
    modelname = EXCLUDED.modelname,
    createdby = EXCLUDED.createdby,
    updateat = EXCLUDED.updateat
WHERE 
    clientmodel.groupid != EXCLUDED.groupid OR
    clientmodel.modelid != EXCLUDED.modelid OR
    clientmodel.groupname != EXCLUDED.groupname OR
    clientmodel.modelname != EXCLUDED.modelname OR
    clientmodel.createdby != EXCLUDED.createdby;
`

	createdBy := "varashebkanthi@intellicar.in"
	updateAt := time.Now().Unix()

	_, err := db.Exec(query,
		pkg.DeviceNo,
		pkg.GroupId,
		pkg.ModelId,
		pkg.GroupNames,
		pkg.Model,
		createdBy,
		updateAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert or update package: %w", err)
	}

	return nil
}

type Groupmodel struct {
	GroupName string
	ModelName string
}

func checkdb(db *sql.DB, Packages map[string]*Clientmodel) ([]*Clientmodel, error) {
	query1 := `SELECT DISTINCT groupname, modelname FROM "clientmodel".clientmodel`
	rows, err := db.Query(query1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groupname and modelname: %w", err)
	}
	defer rows.Close()

	groupModelsMap := make(map[string]struct{})
	for rows.Next() {
		var gm Groupmodel
		if err := rows.Scan(&gm.GroupName, &gm.ModelName); err != nil {
			return nil, fmt.Errorf("failed to scan groupname and modelname: %w", err)
		}
		key := gm.GroupName + "-" + gm.ModelName
		groupModelsMap[key] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over rows: %w", err)
	}

	var missingPackages []*Clientmodel
	for _, pkg := range Packages {
		key := pkg.GroupNames + "-" + pkg.Model
		if _, exists := groupModelsMap[key]; !exists {
			missingPackages = append(missingPackages, pkg)
		}
	}

	return missingPackages, nil
}

func InsertDb(Packages map[string]*Clientmodel) []*Clientmodel {

	dsn := "postgresql://testing_owner:rilHO3obSc7X@ep-weathered-credit-a1p4w5k1.ap-southeast-1.aws.neon.tech/Dmt?sslmode=require"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Fatalf("Database connection failed: %v", err)
	}
	miss, err := checkdb(db, Packages)
	fmt.Println("new client & model to add", len(miss))
	if err != nil {
		fmt.Println("Error: %w", err)
	}

	////
	numWorkers := 10
	packageChan := make(chan *Clientmodel, len(Packages))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(db, packageChan, &wg)
	}

	for _, pkg := range Packages {
		packageChan <- pkg
	}
	close(packageChan)

	wg.Wait()
	fmt.Println("clientmodel successfully inserted into the database.")
	return miss
}

func worker(db *sql.DB, packageChan <-chan *Clientmodel, wg *sync.WaitGroup) {
	defer wg.Done()

	for pkg := range packageChan {
		if err := InsertPackage(db, *pkg); err != nil {
			logger.Printf("Failed to insert Client & model: %v, ERROR: %v", pkg, err)
		}
	}
}
