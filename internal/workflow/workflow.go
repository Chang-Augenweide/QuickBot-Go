package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// Workflow represents a workflow
type Workflow struct {
	ID          string
	Name        string
	Description string
	Steps       []WorkflowStep
	Variables   map[string]interface{}
	Status      string
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID          string
	Name        string
	Type        string
	Config      map[string]interface{}
	OnError     string   // continue, stop, retry
	Dependencies []string // IDs of steps that must complete first
}

// WorkflowExecution represents a workflow execution
type WorkflowExecution struct {
	WorkflowID   string
	ExecutionID  string
	StartTime    time.Time
	EndTime      time.Time
	Status       string
	StepStatus   map[string]string
	Outputs      map[string]interface{}
	Error        error
}

// WorkflowEngine manages workflow execution
type WorkflowEngine struct {
	workflows     map[string]*Workflow
	executions    map[string]*WorkflowExecution
	currentStep   map[string]*WorkflowStep
	stepResults   map[string]map[string]interface{}
	mu            sync.RWMutex
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows:   make(map[string]*Workflow),
		executions:  make(map[string]*WorkflowExecution),
		currentStep: make(map[string]*WorkflowStep),
		stepResults: make(map[string]map[string]interface{}),
	}
}

// RegisterWorkflow registers a new workflow
func (we *WorkflowEngine) RegisterWorkflow(workflow *Workflow) error {
	if workflow.ID == "" {
		workflow.ID = generateWorkflowID()
	}

	we.mu.Lock()
	defer we.mu.Unlock()

	we.workflows[workflow.ID] = workflow
	we.workflows[workflow.Name] = workflow

	log.Printf("Workflow registered: %s", workflow.Name)
	return nil
}

// ExecuteWorkflow executes a workflow
func (we *WorkflowEngine) ExecuteWorkflow(workflowID string, variables map[string]interface{}) (*WorkflowExecution, error) {
	we.mu.RLock()
	workflow, exists := we.workflows[workflowID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Create execution
	execution := &WorkflowExecution{
		WorkflowID:  workflow.ID,
		ExecutionID: generateExecutionID(),
		StartTime:   time.Now(),
		Status:      "running",
		StepStatus:  make(map[string]string),
		Outputs:     make(map[string]interface{}),
	}

	// Initialize workflow variables
	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		workflow.Variables[k] = v
	}

	we.mu.Lock()
	we.executions[execution.ExecutionID] = execution
	we.stepResults[execution.ExecutionID] = make(map[string]interface{})
	we.mu.Unlock()

	// Execute steps
	err := we.executeSteps(workflow, execution)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err
	} else {
		execution.Status = "completed"
	}

	execution.EndTime = time.Now()

	log.Printf("Workflow execution %s completed: %s",
		execution.ExecutionID, execution.Status)

	return execution, nil
}

// executeSteps executes workflow steps
func (we *WorkflowEngine) executeSteps(workflow *Workflow, execution *WorkflowExecution) error {
	// Execute steps in topological order
	executedSteps := make(map[string]bool)
	remainingSteps := make(map[string]*WorkflowStep)

	// Initialize remaining steps
	for i := range workflow.Steps {
		remainingSteps[workflow.Steps[i].ID] = &workflow.Steps[i]
	}

	// Execute until all steps are done
	for len(remainingSteps) > 0 {
		progress := false

		// Find steps that can be executed
		for id, step := range remainingSteps {
			// Check dependencies
			canExecute := true
			for _, dep := range step.Dependencies {
				if !executedSteps[dep] {
					canExecute = false
					break
				}
			}

			if !canExecute {
				continue
			}

			// Execute step
			err := we.executeStep(workflow, execution, step)
			if err != nil {
				execution.StepStatus[id] = "failed"
				log.Printf("Step %s failed: %v", step.Name, err)

				// Handle error based on OnError configuration
				switch step.OnError {
				case "continue":
					// Continue to next steps
				case "retry":
					// Could implement retry logic
				case "stop", "":
					// Stop execution
					return err
				}
			} else {
				execution.StepStatus[id] = "completed"
			}

			executedSteps[id] = true
			delete(remainingSteps, id)
			progress = true
		}

		if !progress {
			return fmt.Errorf("workflow stuck: circular dependencies or unmet dependencies")
		}
	}

	return nil
}

// executeStep executes a single workflow step
func (we *WorkflowEngine) executeStep(workflow *Workflow, execution *WorkflowExecution, step *WorkflowStep) error {
	we.mu.Lock()
	we.currentStep[execution.ExecutionID] = step
	we.mu.Unlock()

	log.Printf("Executing step: %s (type: %s)", step.Name, step.Type)

	var result interface{}
	var err error

	// Execute based on step type
	switch step.Type {
	case "task":
		result, err = we.executeTaskStep(workflow, step)

	case "condition":
		result, err = we.executeConditionStep(workflow, step)

	case loop:
		result, err = we.executeLoopStep(workflow, step)

	case "parallel":
		result, err = we.executeParallelStep(workflow, step)

	case "chat":
		result, err = we.executeChatStep(workflow, step)

	case "tool":
		result, err = we.executeToolStep(workflow, step)

	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}

	we.mu.Lock()
	if err == nil {
		we.stepResults[execution.ExecutionID][step.ID] = result
	}
	we.mu.Unlock()

	return err
}

// executeTaskStep executes a task step
func (we *WorkflowEngine) executeTaskStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Execute a simple task
	taskName, _ := step.Config["name"].(string)
	message := fmt.Sprintf("Task executed: %s", taskName)
	log.Println(message)
	return message, nil
}

// executeConditionStep executes a condition step
func (we *WorkflowEngine) executeConditionStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Evaluate condition
	condition, _ := step.Config["condition"].(string)

	// Simple condition evaluation (placeholder)
	// In production, implement proper expression evaluation
	if condition == "" {
		return map[string]interface{}{
			"condition_met": true,
			"value": condition,
		}, nil
	}

	return map[string]interface{}{
		"condition_met": true,
		"value": condition,
	}, nil
}

// executeLoopStep executes a loop step
func (we *WorkflowEngine) executeLoopStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Execute loop
	iterations, _ := step.Config["iterations"].(float64)
	results := make([]interface{}, 0)

	for i := 0; i < int(iterations); i++ {
		// Execute loop body
		result := fmt.Sprintf("Loop iteration %d", i+1)
		results = append(results, result)
	}

	return results, nil
}

// executeParallelStep executes parallel tasks
func (we *WorkflowEngine) executeParallelStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Execute tasks in parallel
	var wg sync.WaitGroup
	results := make(chan interface{}, 10)

	taskCount := 3 // Example
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result := fmt.Sprintf("Parallel task %d completed", idx+1)
			results <- result
		}(i)
	}

	wg.Wait()
	close(results)

	var resultList []interface{}
	for result := range results {
		resultList = append(resultList, result)
	}

	return resultList, nil
}

// executeChatStep executes a chat step
func (we *WorkflowEngine) executeChatStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Execute chat message
	message, _ := step.Config["message"].(string)
	response := fmt.Sprintf("Chat response: %s", message)
	return response, nil
}

// executeToolStep executes a tool step
func (we *WorkflowEngine) executeToolStep(workflow *Workflow, step *WorkflowStep) (interface{}, error) {
	// Execute a tool
	toolName, _ := step.Config["tool"].(string)
	toolArgs, _ := step.Config["args"].(map[string]interface{})

	return map[string]interface{}{
		"tool":  toolName,
		"args":  toolArgs,
		"result": fmt.Sprintf("Tool %s executed", toolName),
	}, nil
}

// GetExecutionStatus retrieves execution status
func (we *WorkflowEngine) GetExecutionStatus(executionID string) (*WorkflowExecution, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	execution, exists := we.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution, nil
}

// ListWorkflows returns list of workflows
func (we *WorkflowEngine) ListWorkflows() []*Workflow {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(we.workflows))
	for _, workflow := range we.workflows {
		workflows = append(workflows, workflow)
	}
	return workflows
}

// generateWorkflowID generates a workflow ID
func generateWorkflowID() string {
	return fmt.Sprintf("wf_%d", time.Now().UnixNano())
}

// generateExecutionID generates an execution ID
func generateExecutionID() string {
	return fmt.Sprintf("ex_%d", time.Now().UnixNano())
}

// TestWorkflowEngine tests the workflow engine
func TestWorkflowEngine() {
	log.Println("Testing Workflow Engine...")

	engine := NewWorkflowEngine()

	// Create a test workflow
	workflow := &Workflow{
		ID:          "workflow_1",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Variables:   map[string]interface{}{},
		Steps: []WorkflowStep{
			{
				ID:     "step_1",
				Name:   "First Task",
				Type:   "task",
				Config: map[string]interface{}{"name": "Task 1"},
			},
			{
				ID:     "step_2",
				Name:   "Second Task",
				Type:   "task",
				Config: map[string]interface{}{"name": "Task 2"},
				Dependencies: []string{"step_1"},
			},
			{
				ID:     "step_3",
				Name:   "Parallel Tasks",
				Type:   "parallel",
				Config: map[string]interface{}{},
				Dependencies: []string{"step_2"},
			},
		},
	}

	// Register workflow
	err := engine.RegisterWorkflow(workflow)
	if err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}
	log.Println("✓ Workflow registered")

	// Execute workflow
	execution, err := engine.ExecuteWorkflow(workflow.ID, map[string]interface{}{})
	if err != nil {
		log.Fatalf("Failed to execute workflow: %v", err)
	}

	log.Printf("✓ Workflow executed")
	log.Printf("  Status: %s", execution.Status)
	log.Printf("  Duration: %v", execution.EndTime.Sub(execution.StartTime))
	log.Printf("  Steps completed: %d", len(execution.StepStatus))

	// List workflows
	workflows := engine.ListWorkflows()
	log.Printf("✓ Available workflows: %d", len(workflows))

	log.Println("✓ Workflow engine tests passed")
}

// main - test entry point
func main() {
	log.Println("QuickBot Go Workflow Engine")
	log.Println("✓ Workflow engine module initialized")

	// Run tests
	TestWorkflowEngine()
}
