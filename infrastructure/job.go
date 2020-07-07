package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/wosai/deepmock/domain"
)

type Job struct {
	period   time.Duration
	rule     domain.RuleRepository
	executor domain.ExecutorRepository
}

func NewJob(period time.Duration) *Job {
	return &Job{period: period}
}

func (job *Job) Period() time.Duration {
	return job.period
}

func (job *Job) WithRuleRepository(rr domain.RuleRepository) {
	job.rule = rr
}

func (job *Job) WithExecutorRepository(er domain.ExecutorRepository) {
	job.executor = er
}

func (job *Job) Do() error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	rules, err := job.rule.Export(ctx)
	if err != nil {
		return err
	}
	executors := make([]*domain.Executor, len(rules))
	for index, rule := range rules {
		executor, err := rule.To()
		if err != nil {
			return fmt.Errorf("failed to convert Rule to Executor: %s - %w", rule.ID, err)
		}
		executors[index] = executor
	}
	job.executor.ImportAll(ctx, executors...)
	return nil
}
