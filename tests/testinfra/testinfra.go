//go:build integration

// Использование:
//
//	//go:build integration
//
//	func TestSomething(t *testing.T) {
//	    pg := testinfra.NewPostgres(t, filepath.Join(testinfra.ProjectRoot(t), "migrations/auth"))
//	    txm := transaction.New(pg.Pool)
//	    repo := user.New(txm)
//	    // ... test with real Postgres
//	}
//
// запустить интеграционные тетсы:
//
//	go test -tags=integration -v -count=1 ./tests/...
package testinfra
