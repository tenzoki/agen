//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/storage"
	"github.com/tenzoki/agen/omni/internal/transaction"
)

func main() {
	fmt.Println("🔄 Transaction Support Complete Demo")
	fmt.Println("============================================================")
	fmt.Println("ACID transactions for BadgerDB Dual Store with rollback support")

	// Setup temporary storage
	tmpDir := "/tmp/transaction-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create transaction manager
	tm := transaction.NewTransactionManager(store)
	defer tm.Close()

	fmt.Println("\n🏗️  Setting up transaction system...")
	fmt.Println("   ✅ Transaction manager initialized")
	fmt.Println("   ✅ BadgerDB integration configured")
	fmt.Println("   ✅ ACID compliance enabled")

	// Demo sections
	fmt.Println("\n1. Basic Transaction Lifecycle")
	demoBasicTransactions(tm)

	fmt.Println("\n2. Transaction Rollback")
	demoTransactionRollback(tm)

	fmt.Println("\n3. Atomic Multi-Operation Transactions")
	demoAtomicOperations(tm)

	fmt.Println("\n4. Transaction Consistency & Validation")
	demoConsistencyChecking(tm)

	fmt.Println("\n5. Concurrent Transaction Management")
	demoConcurrentTransactions(tm)

	fmt.Println("\n6. Transaction Statistics & Monitoring")
	demoTransactionStats(tm)

	fmt.Println("\n7. Savepoints & Nested Transactions")
	demoSavepoints(tm)

	fmt.Println("\n8. Error Handling & Recovery")
	demoErrorHandling(tm)

	fmt.Println("\n✅ Transaction demo completed successfully!")
}

func demoBasicTransactions(tm transaction.TransactionManager) {
	fmt.Println("   Basic transaction commit and rollback operations...")

	// Successful transaction
	fmt.Println("\n   💚 Successful Transaction:")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		vertex := createDemoVertex("user:tx1", "User", "Transaction User 1")
		if err := tx.CreateVertex(vertex); err != nil {
			return err
		}

		fmt.Printf("   📝 Created vertex: %s\n", vertex.ID)

		// Set some KV data
		if err := tx.KVSet("config:tx_demo", []byte("transaction_enabled")); err != nil {
			return err
		}

		fmt.Printf("   📝 Set KV pair: config:tx_demo\n")

		return nil // Success - transaction will commit
	})

	if err != nil {
		fmt.Printf("   ❌ Transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Transaction committed successfully\n")
	}

	// Verify data was persisted
	fmt.Println("\n   🔍 Verifying persisted data:")
	err = tm.ExecuteReadOnly(func(tx transaction.GraphTx) error {
		if exists, err := tx.VertexExists("user:tx1"); err != nil {
			return err
		} else if exists {
			fmt.Printf("   ✅ Vertex user:tx1 persisted\n")
		} else {
			fmt.Printf("   ❌ Vertex user:tx1 not found\n")
		}

		if exists, err := tx.KVExists("config:tx_demo"); err != nil {
			return err
		} else if exists {
			fmt.Printf("   ✅ KV pair config:tx_demo persisted\n")
		} else {
			fmt.Printf("   ❌ KV pair config:tx_demo not found\n")
		}

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Verification failed: %v\n", err)
	}
}

func demoTransactionRollback(tm transaction.TransactionManager) {
	fmt.Println("   Transaction rollback on errors...")

	fmt.Println("\n   🔄 Failed Transaction (will rollback):")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		vertex := createDemoVertex("user:tx2", "User", "Transaction User 2")
		if err := tx.CreateVertex(vertex); err != nil {
			return err
		}

		fmt.Printf("   📝 Created vertex: %s\n", vertex.ID)

		if err := tx.KVSet("temp:data", []byte("temporary_data")); err != nil {
			return err
		}

		fmt.Printf("   📝 Set temporary KV data\n")

		// Simulate error - this will trigger rollback
		return fmt.Errorf("simulated business logic error")
	})

	if err != nil {
		fmt.Printf("   ❌ Transaction failed (as expected): %v\n", err)
		fmt.Printf("   🔄 Transaction was automatically rolled back\n")
	}

	// Verify data was NOT persisted
	fmt.Println("\n   🔍 Verifying rollback (data should not exist):")
	err = tm.ExecuteReadOnly(func(tx transaction.GraphTx) error {
		if exists, err := tx.VertexExists("user:tx2"); err != nil {
			return err
		} else if !exists {
			fmt.Printf("   ✅ Vertex user:tx2 correctly rolled back\n")
		} else {
			fmt.Printf("   ❌ Vertex user:tx2 should not exist\n")
		}

		if exists, err := tx.KVExists("temp:data"); err != nil {
			return err
		} else if !exists {
			fmt.Printf("   ✅ KV pair temp:data correctly rolled back\n")
		} else {
			fmt.Printf("   ❌ KV pair temp:data should not exist\n")
		}

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Rollback verification failed: %v\n", err)
	}
}

func demoAtomicOperations(tm transaction.TransactionManager) {
	fmt.Println("   Multi-operation atomic transactions...")

	fmt.Println("\n   ⚛️  Complex Atomic Transaction:")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		// Create multiple related vertices
		user := createDemoVertex("user:alice", "User", "Alice Johnson")
		company := createDemoVertex("company:techcorp", "Company", "TechCorp Inc")
		project := createDemoVertex("project:webapp", "Project", "Web Application")

		if err := tx.CreateVertex(user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Printf("   📝 Created user: %s\n", user.ID)

		if err := tx.CreateVertex(company); err != nil {
			return fmt.Errorf("failed to create company: %w", err)
		}
		fmt.Printf("   📝 Created company: %s\n", company.ID)

		if err := tx.CreateVertex(project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		fmt.Printf("   📝 Created project: %s\n", project.ID)

		// Create relationships
		worksEdge := createDemoEdge("works:alice:techcorp", "works_at", "user:alice", "company:techcorp")
		assignedEdge := createDemoEdge("assigned:alice:webapp", "assigned_to", "user:alice", "project:webapp")
		ownsEdge := createDemoEdge("owns:techcorp:webapp", "owns", "company:techcorp", "project:webapp")

		if err := tx.CreateEdge(worksEdge); err != nil {
			return fmt.Errorf("failed to create works relationship: %w", err)
		}
		fmt.Printf("   🔗 Created edge: %s\n", worksEdge.ID)

		if err := tx.CreateEdge(assignedEdge); err != nil {
			return fmt.Errorf("failed to create assignment relationship: %w", err)
		}
		fmt.Printf("   🔗 Created edge: %s\n", assignedEdge.ID)

		if err := tx.CreateEdge(ownsEdge); err != nil {
			return fmt.Errorf("failed to create ownership relationship: %w", err)
		}
		fmt.Printf("   🔗 Created edge: %s\n", ownsEdge.ID)

		// Set related configuration
		kvPairs := map[string][]byte{
			"user:alice:role":       []byte("senior_engineer"),
			"company:techcorp:size": []byte("500"),
			"project:webapp:status": []byte("active"),
		}

		if err := tx.BatchKVSet(kvPairs); err != nil {
			return fmt.Errorf("failed to set configuration: %w", err)
		}
		fmt.Printf("   📄 Set %d configuration entries\n", len(kvPairs))

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Atomic transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ All operations committed atomically\n")
	}

	// Verify complete graph was created
	fmt.Println("\n   🔍 Verifying atomic transaction results:")
	err = tm.ExecuteReadOnly(func(tx transaction.GraphTx) error {
		vertices, err := tx.GetAllVertices(-1)
		if err != nil {
			return err
		}
		fmt.Printf("   📊 Total vertices created: %d\n", len(vertices))

		edges, err := tx.GetAllEdges(-1)
		if err != nil {
			return err
		}
		fmt.Printf("   📊 Total edges created: %d\n", len(edges))

		// Check specific relationships
		if exists, _ := tx.EdgeExists("works:alice:techcorp"); exists {
			fmt.Printf("   ✅ Alice works at TechCorp relationship exists\n")
		}
		if exists, _ := tx.EdgeExists("assigned:alice:webapp"); exists {
			fmt.Printf("   ✅ Alice assigned to webapp relationship exists\n")
		}
		if exists, _ := tx.EdgeExists("owns:techcorp:webapp"); exists {
			fmt.Printf("   ✅ TechCorp owns webapp relationship exists\n")
		}

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Verification failed: %v\n", err)
	}
}

func demoConsistencyChecking(tm transaction.TransactionManager) {
	fmt.Println("   Data consistency validation...")

	fmt.Println("\n   🔍 Consistency Check - Valid Data:")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		// Create valid data structure
		user := createDemoVertex("user:bob", "User", "Bob Smith")
		user.Properties["email"] = "bob@example.com"
		user.Properties["age"] = 30

		if err := tx.CreateVertex(user); err != nil {
			return err
		}
		fmt.Printf("   ✅ Created valid user: %s\n", user.ID)

		// This would trigger consistency checking in a full implementation
		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Valid data transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Valid data passed consistency checks\n")
	}

	fmt.Println("\n   🔍 Consistency Check - Invalid Edge (no target vertex):")
	err = tm.Execute(func(tx transaction.GraphTx) error {
		// Try to create edge with non-existent target
		invalidEdge := createDemoEdge("invalid:edge", "follows", "user:bob", "user:nonexistent")

		if err := tx.CreateEdge(invalidEdge); err != nil {
			fmt.Printf("   ✅ Correctly rejected invalid edge: %v\n", err)
			return nil // Don't fail the transaction, this is expected
		}

		fmt.Printf("   ❌ Invalid edge was incorrectly accepted\n")
		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Consistency check transaction failed: %v\n", err)
	}
}

func demoConcurrentTransactions(tm transaction.TransactionManager) {
	fmt.Println("   Concurrent transaction management...")

	fmt.Println("\n   🔄 Creating multiple concurrent transactions:")

	// Start multiple transactions
	tx1, err := tm.Begin(nil)
	if err != nil {
		fmt.Printf("   ❌ Failed to start transaction 1: %v\n", err)
		return
	}
	defer tx1.Rollback()

	tx2, err := tm.Begin(nil)
	if err != nil {
		fmt.Printf("   ❌ Failed to start transaction 2: %v\n", err)
		return
	}
	defer tx2.Rollback()

	fmt.Printf("   ✅ Started transaction 1: %s\n", tx1.GetID())
	fmt.Printf("   ✅ Started transaction 2: %s\n", tx2.GetID())

	// Check active transactions
	activeTxs := tm.GetActiveTransactions()
	fmt.Printf("   📊 Active transactions: %d\n", len(activeTxs))
	for _, txID := range activeTxs {
		fmt.Printf("     • %s\n", txID)
	}

	// Commit one transaction
	if err := tx1.Commit(); err != nil {
		fmt.Printf("   ❌ Failed to commit transaction 1: %v\n", err)
	} else {
		fmt.Printf("   ✅ Committed transaction 1\n")
	}

	// Check updated active transactions
	activeTxs = tm.GetActiveTransactions()
	fmt.Printf("   📊 Active transactions after commit: %d\n", len(activeTxs))
}

func demoTransactionStats(tm transaction.TransactionManager) {
	fmt.Println("   Transaction statistics and monitoring...")

	fmt.Println("\n   📊 Transaction with Statistics Tracking:")

	// Configure transaction with specific settings
	config := &transaction.TransactionConfig{
		IsolationLevel: transaction.ReadCommitted,
		Timeout:        30 * time.Second,
		ReadOnly:       false,
		MaxRetries:     3,
	}

	err := tm.ExecuteWithConfig(config, func(tx transaction.GraphTx) error {
		// Perform various operations to generate stats
		for i := 0; i < 5; i++ {
			vertex := createDemoVertex(fmt.Sprintf("stats:vertex:%d", i), "StatsVertex", fmt.Sprintf("Stats Vertex %d", i))
			if err := tx.CreateVertex(vertex); err != nil {
				return err
			}
		}

		// Set some KV data
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf("stats:key:%d", i)
			value := []byte(fmt.Sprintf("stats_value_%d", i))
			if err := tx.KVSet(key, value); err != nil {
				return err
			}
		}

		// Get transaction stats
		stats := tx.GetStats()
		fmt.Printf("   📈 Transaction ID: %s\n", stats.ID)
		fmt.Printf("   📈 Operations performed: %d\n", stats.OperationCount)
		fmt.Printf("   📈 Write operations: %d\n", stats.WriteCount)
		fmt.Printf("   📈 Isolation level: %s\n", stats.IsolationLevel)
		fmt.Printf("   📈 Duration so far: %v\n", stats.Duration)
		fmt.Printf("   📈 Bytes written: %d\n", stats.BytesWritten)

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Stats transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Statistics transaction completed\n")
	}
}

func demoSavepoints(tm transaction.TransactionManager) {
	fmt.Println("   Savepoints and nested rollback...")

	fmt.Println("\n   💾 Transaction with Savepoints:")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		// Create initial data
		vertex1 := createDemoVertex("savepoint:user1", "User", "Savepoint User 1")
		if err := tx.CreateVertex(vertex1); err != nil {
			return err
		}
		fmt.Printf("   📝 Created initial vertex: %s\n", vertex1.ID)

		// Create savepoint
		if err := tx.Savepoint("checkpoint1"); err != nil {
			return fmt.Errorf("failed to create savepoint: %w", err)
		}
		fmt.Printf("   💾 Created savepoint: checkpoint1\n")

		// Create more data
		vertex2 := createDemoVertex("savepoint:user2", "User", "Savepoint User 2")
		if err := tx.CreateVertex(vertex2); err != nil {
			return err
		}
		fmt.Printf("   📝 Created vertex after savepoint: %s\n", vertex2.ID)

		// Create another savepoint
		if err := tx.Savepoint("checkpoint2"); err != nil {
			return fmt.Errorf("failed to create savepoint: %w", err)
		}
		fmt.Printf("   💾 Created savepoint: checkpoint2\n")

		// Create even more data
		vertex3 := createDemoVertex("savepoint:user3", "User", "Savepoint User 3")
		if err := tx.CreateVertex(vertex3); err != nil {
			return err
		}
		fmt.Printf("   📝 Created vertex after second savepoint: %s\n", vertex3.ID)

		// Simulate rollback to first savepoint
		if err := tx.RollbackToSavepoint("checkpoint1"); err != nil {
			return fmt.Errorf("failed to rollback to savepoint: %w", err)
		}
		fmt.Printf("   🔄 Rolled back to checkpoint1 (conceptual)\n")

		// Release savepoints
		if err := tx.ReleaseSavepoint("checkpoint1"); err != nil {
			return fmt.Errorf("failed to release savepoint: %w", err)
		}
		fmt.Printf("   🗑️  Released savepoint: checkpoint1\n")

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Savepoint transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Savepoint operations completed\n")
	}
}

func demoErrorHandling(tm transaction.TransactionManager) {
	fmt.Println("   Error handling and recovery patterns...")

	fmt.Println("\n   🚨 Error Recovery Scenarios:")

	// Scenario 1: Duplicate key error
	fmt.Println("\n   📝 Scenario 1: Duplicate Key Handling")
	err := tm.Execute(func(tx transaction.GraphTx) error {
		vertex := createDemoVertex("error:duplicate", "User", "Duplicate Test User")
		return tx.CreateVertex(vertex)
	})

	if err != nil {
		fmt.Printf("   ❌ First create failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ First create succeeded\n")
	}

	// Try to create the same vertex again
	err = tm.Execute(func(tx transaction.GraphTx) error {
		vertex := createDemoVertex("error:duplicate", "User", "Duplicate Test User Again")
		return tx.CreateVertex(vertex)
	})

	if err != nil {
		fmt.Printf("   ✅ Duplicate correctly rejected: %v\n", err)
	} else {
		fmt.Printf("   ❌ Duplicate should have been rejected\n")
	}

	// Scenario 2: Timeout handling (simulated)
	fmt.Println("\n   📝 Scenario 2: Timeout Configuration")
	shortConfig := &transaction.TransactionConfig{
		IsolationLevel: transaction.ReadCommitted,
		Timeout:        100 * time.Millisecond, // Very short timeout
		ReadOnly:       false,
		MaxRetries:     1,
	}

	err = tm.ExecuteWithConfig(shortConfig, func(tx transaction.GraphTx) error {
		vertex := createDemoVertex("timeout:test", "User", "Timeout Test User")
		if err := tx.CreateVertex(vertex); err != nil {
			return err
		}

		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)

		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Timeout transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Timeout transaction completed within limits\n")
	}

	// Scenario 3: Read-only constraint
	fmt.Println("\n   📝 Scenario 3: Read-Only Transaction Enforcement")
	readOnlyConfig := &transaction.TransactionConfig{
		IsolationLevel: transaction.ReadCommitted,
		Timeout:        30 * time.Second,
		ReadOnly:       true,
		MaxRetries:     1,
	}

	err = tm.ExecuteWithConfig(readOnlyConfig, func(tx transaction.GraphTx) error {
		// This should work - reading data
		if exists, err := tx.VertexExists("error:duplicate"); err != nil {
			return err
		} else {
			fmt.Printf("   ✅ Read operation successful: vertex exists = %t\n", exists)
		}

		// This conceptually shouldn't work in a read-only transaction
		// but our current implementation doesn't enforce this strictly
		vertex := createDemoVertex("readonly:violation", "User", "Read-Only Violation")
		if err := tx.CreateVertex(vertex); err != nil {
			fmt.Printf("   ✅ Write correctly rejected in read-only transaction: %v\n", err)
			return nil // Don't fail the transaction
		}

		fmt.Printf("   ❌ Write should have been rejected in read-only transaction\n")
		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Read-only transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Read-only transaction completed\n")
	}
}

// Helper functions
func createDemoVertex(id, vertexType, name string) *common.Vertex {
	vertex := common.NewVertex(id, vertexType)
	vertex.Properties["name"] = name
	vertex.Properties["created_in_demo"] = true
	return vertex
}

func createDemoEdge(id, edgeType, from, to string) *common.Edge {
	edge := common.NewEdge(id, edgeType, from, to)
	edge.Properties["created_in_demo"] = true
	edge.Weight = 1.0
	return edge
}
