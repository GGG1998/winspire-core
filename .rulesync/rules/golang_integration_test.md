---
targets:
  - '*'
root: false
globs:
  - '*integration_test.go'
  - '*test.go'
cursor:
  alwaysApply: false
  globs:
    - '*integration_test.go'
    - '*test.go'
---
# RULES
- create folder fixtures if it needs
- Fixture folder COULD HAVE:
    - raw sql query
    - curl collection
- Don't implement helper function that already exists 
- for repeating tasks create table-driven test

**BAD:**
```
// calculator_test.go - BEZ table-driven
package calculator

import "testing"

// Osobne funkcje dla każdego przypadku
func TestAddPositiveNumbers(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}

func TestAddNegativeNumbers(t *testing.T) {
    result := Add(-2, -3)
    if result != -5 {
        t.Errorf("Add(-2, -3) = %d; want -5", result)
    }
}

func TestAddWithZero(t *testing.T) {
    result := Add(5, 0)
    if result != 5 {
        t.Errorf("Add(5, 0) = %d; want 5", result)
    }
}

func TestAddMixedNumbers(t *testing.T) {
    result := Add(-2, 5)
    if result != 3 {
        t.Errorf("Add(-2, 5) = %d; want 3", result)
    }
}

// Albo wszystko w jednej funkcji - dużo powtórzeń
func TestDivide(t *testing.T) {
    result1, err1 := Divide(10, 2)
    if err1 != nil {
        t.Errorf("unexpected error: %v", err1)
    }
    if result1 != 5 {
        t.Errorf("Divide(10, 2) = %d; want 5", result1)
    }

    result2, err2 := Divide(9, 3)
    if err2 != nil {
        t.Errorf("unexpected error: %v", err2)
    }
    if result2 != 3 {
        t.Errorf("Divide(9, 3) = %d; want 3", result2)
    }

    _, err3 := Divide(10, 0)
    if err3 == nil {
        t.Error("expected error for division by zero")
    }
}
```

**GOOD:**
```go
package calculator

import "testing"

func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"with zero", 5, 0, 5},
        {"mixed numbers", -2, 5, 3},
        {"large numbers", 1000000, 2000000, 3000000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

func TestDivide(t *testing.T) {
    tests := []struct {
        name        string
        a, b        int
        expected    int
        expectError bool
    }{
        {"normal division", 10, 2, 5, false},
        {"exact division", 9, 3, 3, false},
        {"division by zero", 10, 0, 0, true},
        {"negative dividend", -10, 2, -5, false},
        {"negative divisor", 10, -2, -5, false},
        {"both negative", -10, -2, 5, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)
            
            if tt.expectError {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

func TestIsEven(t *testing.T) {
    tests := []struct {
        input    int
        expected bool
    }{
        {2, true},
        {3, false},
        {0, true},
        {-4, true},
        {-3, false},
        {100, true},
        {101, false},
    }

    for _, tt := range tests {
        t.Run(fmt.Sprintf("IsEven(%d)", tt.input), func(t *testing.T) {
            if result := IsEven(tt.input); result != tt.expected {
                t.Errorf("got %v; want %v", result, tt.expected)
            }
        })
    }
}
```

## Fixture structure example:

```go
func LoadSQLFixtures(cfg config.Config) error {
	pathToFile := "/app/fixtures.sql"
	q, err := os.ReadFile(pathToFile)
	if err != nil {
		return errors.WithStack(err)
	}

	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = db.Exec(string(q))
	if err != nil {
		return errors.WithStack(err)
	}
	err = db.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
```

# STRUCTURE 

## TEST MAIN

TestMain is a special function that gets called before all other tests in the package it's located in. Here, we're being (justifiable, I would say) paranoid and call a function that drops everything in the database so we are sure we are starting from a clean slate. You can find the function in repository/psql.go if you want to take a closer look.

```go
// TestMain gets run before running any other _test.go files in each package
// here, we use it to make sure we start from a clean slate
func TestMain(m *testing.M) {
	cfg := config.NewConfig()
	// make sure we start from a clean slate
	err := psql.DropEverythingInDatabase(*cfg)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
```

## RUN MIGRATION

Avoid putting code directly to the test, we are able to load migration from files.

**Bad:**
```go
_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS prelobbies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL UNIQUE,
			status VARCHAR(30) NOT NULL DEFAULT 'waiting' CHECK (
				status IN ('waiting', 'grace_period', 'generating_bracket', 'started', 'cancelled')
			),
			grace_period_start TIMESTAMP,
			grace_period_end TIMESTAMP,
			min_participants INTEGER NOT NULL DEFAULT 2 CHECK (min_participants >= 2),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}

	// Create prelobby_activity_feed table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS prelobby_activity_feed (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tournament_id UUID NOT NULL REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
			event_type VARCHAR(50) NOT NULL CHECK (
				event_type IN ('participant_joined', 'participant_left', 'grace_period_started',
							   'grace_period_ended', 'bracket_generation', 'tournament_cancelled', 'system_message')
			),
			message TEXT NOT NULL,
			participant_name VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}
```

**Good:**
```go
func RunUpMigrations(cfg config.Config) error {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(b), "../migrations")
	migrationDir := filepath.Join("file://" + basePath)
	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.WithStack(err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(migrationDir, "postgres", driver)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, ErrNoNewMigrations) {
			return errors.WithStack(err)
		}
	}
	m.Close()
	return nil
}
```

## DOWN MIGRATION

```go
func RunDownMigrations(cfg config.Config) error {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(b), "../migrations")
	migrationDir := filepath.Join("file://" + basePath)
	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.WithStack(err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(migrationDir, "postgres", driver)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := m.Down(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
```
