package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/wosai/deepmock/domain"
)

// Job AsyncJob的实现
type Job struct {
	period   time.Duration
	rule     domain.RuleRepository
	executor domain.ExecutorRepository
}

// NewJob 工厂函数
func NewJob(period time.Duration) *Job {
	return &Job{period: period}
}

// Period 执行周期
func (job *Job) Period() time.Duration {
	return job.period
}

// WithRuleRepository 载入规则存储库
func (job *Job) WithRuleRepository(rr domain.RuleRepository) {
	job.rule = rr
}

// WithExecutorRepository 载入执行器存储库
func (job *Job) WithExecutorRepository(er domain.ExecutorRepository) {
	job.executor = er
}

// Do 任务逻辑
func (job *Job) Do() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
