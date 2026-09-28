package planservices

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	emailservices "goravel/app/services/email"
	"goravel/app/services/notifications"
	settingsservices "goravel/app/services/settings"
)

const (
	DefaultExpiryGracePeriodDays = 3
	DefaultExpiryReminderDays    = "7,3,1"
	DefaultFallbackPlanCode      = "free"
	DefaultOverQuotaPolicy       = "keep_data_block_upload"
)

type ExpiryPolicy struct {
	FallbackPlanCode string
	GracePeriodDays  int
	ReminderDays     []int
	EmailEnabled     bool
	OverQuotaPolicy  string
}

type LifecycleReport struct {
	Scanned            int
	Expired            int
	RemindersScheduled int
	EmailsSent         int
	EmailsDeferred     int
	EmailFailures      int
}

type SubscriptionLifecycleService struct {
	settings *settingsservices.SettingService
	email    *emailservices.Service
	notify   *notifications.NotificationService
}

func NewSubscriptionLifecycleService() *SubscriptionLifecycleService {
	return &SubscriptionLifecycleService{
		settings: settingsservices.NewSettingService(),
		email:    emailservices.NewService(),
		notify:   notifications.NewNotificationService(),
	}
}

func (s *SubscriptionLifecycleService) Policy() ExpiryPolicy {
	return expiryPolicy(s.settings)
}

func expiryPolicy(settings *settingsservices.SettingService) ExpiryPolicy {
	if settings == nil {
		settings = settingsservices.NewSettingService()
	}
	graceDays := expirySettingInt(settings.Resolve("subscription.expiry.grace_period_days", ""), DefaultExpiryGracePeriodDays)
	if graceDays > 30 {
		graceDays = 30
	}
	reminderDays := normalizeExpiryReminderDays(settings.Resolve("subscription.expiry.reminder_days", DefaultExpiryReminderDays))
	if len(reminderDays) == 0 {
		reminderDays = normalizeExpiryReminderDays(DefaultExpiryReminderDays)
	}
	fallback := strings.TrimSpace(settings.Resolve("subscription.expiry.fallback_plan_code", DefaultFallbackPlanCode))
	if fallback == "" {
		fallback = DefaultFallbackPlanCode
	}
	quotaPolicy := strings.TrimSpace(settings.Resolve("subscription.expiry.over_quota_policy", DefaultOverQuotaPolicy))
	if quotaPolicy == "" {
		quotaPolicy = DefaultOverQuotaPolicy
	}
	return ExpiryPolicy{
		FallbackPlanCode: fallback,
		GracePeriodDays:  graceDays,
		ReminderDays:     reminderDays,
		EmailEnabled:     expirySettingBool(settings.Resolve("subscription.expiry.email_enabled", ""), true),
		OverQuotaPolicy:  quotaPolicy,
	}
}

func expirySettingInt(raw string, fallback int) int {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func expirySettingBool(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

func (s *SubscriptionLifecycleService) Process(now time.Time) (LifecycleReport, error) {
	policy := s.Policy()
	var subscriptions []models.Subscription
	if err := facades.Orm().Query().WhereIn("status", []any{subscriptionStateActive, subscriptionStateExpired}).WhereNotNull("ends_at").Get(&subscriptions); err != nil {
		return LifecycleReport{}, err
	}
	report := LifecycleReport{Scanned: len(subscriptions)}
	for index := range subscriptions {
		result, err := s.processSubscription(&subscriptions[index], now.UTC(), policy)
		if err != nil {
			return report, err
		}
		report.Expired += result.Expired
		report.RemindersScheduled += result.RemindersScheduled
		report.EmailsSent += result.EmailsSent
		report.EmailsDeferred += result.EmailsDeferred
		report.EmailFailures += result.EmailFailures
	}
	return report, nil
}

func (s *SubscriptionLifecycleService) processSubscription(subscription *models.Subscription, now time.Time, policy ExpiryPolicy) (LifecycleReport, error) {
	report := LifecycleReport{}
	if subscription == nil || subscription.EndsAt == nil {
		return report, nil
	}
	var user models.User
	if err := facades.Orm().Query().Where("id = ?", subscription.UserID).First(&user); err != nil {
		return report, err
	}
	var plan models.Plan
	if err := facades.Orm().Query().Where("id = ?", subscription.PlanID).First(&plan); err != nil {
		return report, err
	}
	endsAt := subscription.EndsAt.UTC()
	if subscription.Status == subscriptionStateActive && !now.Before(endsAt) {
		changed, err := s.expireSubscription(subscription, now, policy)
		if err != nil {
			return report, err
		}
		if changed {
			report.Expired++
		}
		subscription.Status = subscriptionStateExpired
	}

	if now.Before(endsAt) && subscription.Status == subscriptionStateActive {
		for _, days := range policy.ReminderDays {
			if now.Before(endsAt.Add(-time.Duration(days) * 24 * time.Hour)) {
				continue
			}
			result, err := s.processEvent(subscription, &user, &plan, fmt.Sprintf("%dd", days), now, policy)
			if err != nil {
				return report, err
			}
			report.RemindersScheduled += result.RemindersScheduled
			report.EmailsSent += result.EmailsSent
			report.EmailsDeferred += result.EmailsDeferred
			report.EmailFailures += result.EmailFailures
		}
		return report, nil
	}

	result, err := s.processEvent(subscription, &user, &plan, "expired", now, policy)
	if err != nil {
		return report, err
	}
	report.RemindersScheduled += result.RemindersScheduled
	report.EmailsSent += result.EmailsSent
	report.EmailsDeferred += result.EmailsDeferred
	report.EmailFailures += result.EmailFailures
	return report, nil
}

func (s *SubscriptionLifecycleService) processEvent(subscription *models.Subscription, user *models.User, plan *models.Plan, kind string, now time.Time, policy ExpiryPolicy) (LifecycleReport, error) {
	report := LifecycleReport{}
	key := expiryNotificationKey(subscription.ID, kind)
	delivery, created, err := ensureExpiryDelivery(subscription, kind, key, now)
	if err != nil {
		return report, err
	}
	if created {
		report.RemindersScheduled++
		s.publishInAppNotification(user, plan, kind, subscription.EndsAt)
	}
	if delivery == nil || delivery.Status == "sent" {
		return report, nil
	}
	result, err := s.deliverEmail(delivery, user, plan, kind, now, policy)
	if err != nil {
		return report, err
	}
	report.EmailsSent += result.EmailsSent
	report.EmailsDeferred += result.EmailsDeferred
	report.EmailFailures += result.EmailFailures
	return report, nil
}

func ensureExpiryDelivery(subscription *models.Subscription, kind, key string, now time.Time) (*models.SubscriptionNotificationDelivery, bool, error) {
	var delivery models.SubscriptionNotificationDelivery
	exists, err := facades.Orm().Query().Model(&models.SubscriptionNotificationDelivery{}).Where("event_key = ?", key).Exists()
	if err != nil {
		return nil, false, err
	}
	created := false
	if !exists {
		if err := facades.Orm().Query().Create(&models.SubscriptionNotificationDelivery{
			SubscriptionID: subscription.ID,
			UserID:         subscription.UserID,
			EventKey:       key,
			Kind:           kind,
			Channel:        "email",
			Status:         "pending",
			NextAttemptAt:  &now,
		}); err != nil {
			return nil, false, err
		}
		created = true
	}
	if err := facades.Orm().Query().Where("event_key = ?", key).First(&delivery); err != nil {
		return nil, false, err
	}
	return &delivery, created, nil
}

func (s *SubscriptionLifecycleService) expireSubscription(subscription *models.Subscription, now time.Time, policy ExpiryPolicy) (bool, error) {
	changed := false
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var current models.Subscription
		if err := tx.Where("id = ?", subscription.ID).First(&current); err != nil {
			return err
		}
		if current.Status == subscriptionStateActive && current.EndsAt != nil && !current.EndsAt.After(now) {
			graceEnds := gracePeriodEndsAt(current.EndsAt.UTC(), policy.GracePeriodDays)
			if _, err := tx.Where("id = ?").Update(map[string]any{
				"status": "expired", "grace_period_ends_at": graceEnds, "updated_at": now,
			}); err != nil {
				return err
			}
			changed = true
		}
		return ensureFallbackSubscription(tx, current.UserID, policy.FallbackPlanCode, now)
	})
	return changed, err
}

func ensureFallbackSubscription(query orm.Query, userID uint, code string, now time.Time) error {
	exists, err := query.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, subscriptionStateActive).Exists()
	if err != nil || exists {
		return err
	}
	var plan models.Plan
	if err := query.Where("code = ? AND status = ?", code, "active").First(&plan); err != nil && code != DefaultFallbackPlanCode {
		if err := query.Where("code = ? AND status = ?", DefaultFallbackPlanCode, "active").First(&plan); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if strings.TrimSpace(plan.EntitlementsJSON) == "" {
		return ErrPlanNotFound
	}
	snapshot, err := BuildSubscriptionSnapshot(plan)
	if err != nil {
		return err
	}
	return CreateSubscriptionWithQuery(query, models.Subscription{
		UserID: userID, PlanID: plan.ID, Status: subscriptionStateActive,
		StartsAt: &now, EntitlementSnapshotJSON: snapshot,
	})
}

type emailDeliveryResult struct {
	EmailsSent     int
	EmailsDeferred int
	EmailFailures  int
}

func (s *SubscriptionLifecycleService) deliverEmail(delivery *models.SubscriptionNotificationDelivery, user *models.User, plan *models.Plan, kind string, now time.Time, policy ExpiryPolicy) (emailDeliveryResult, error) {
	if delivery.NextAttemptAt != nil && delivery.NextAttemptAt.After(now) {
		return emailDeliveryResult{}, nil
	}
	if !policy.EmailEnabled {
		next := now.Add(24 * time.Hour)
		_, err := facades.Orm().Query().Where("id = ?", delivery.ID).Update(map[string]any{"status": "deferred", "next_attempt_at": next, "last_error": "subscription expiry email disabled"})
		return emailDeliveryResult{EmailsDeferred: 1}, err
	}
	message := expiryEmailMessage(user, plan, kind, delivery.SubscriptionID)
	err := s.email.Send(context.Background(), message)
	if err == nil {
		sentAt := now
		_, updateErr := facades.Orm().Query().Where("id = ?", delivery.ID).Update(map[string]any{"status": "sent", "sent_at": sentAt, "next_attempt_at": nil, "last_error": nil, "updated_at": now})
		return emailDeliveryResult{EmailsSent: 1}, updateErr
	}
	attempts := delivery.Attempts + 1
	next := now.Add(emailRetryDelay(attempts))
	status := "pending"
	if errors.Is(err, emailservices.ErrEmailDisabled) || errors.Is(err, emailservices.ErrEmailNotConfigured) {
		status = "deferred"
	}
	_, updateErr := facades.Orm().Query().Where("id = ?", delivery.ID).Update(map[string]any{"status": status, "attempts": attempts, "next_attempt_at": next, "last_error": err.Error(), "updated_at": now})
	if updateErr != nil {
		return emailDeliveryResult{}, updateErr
	}
	if status == "deferred" {
		return emailDeliveryResult{EmailsDeferred: 1}, nil
	}
	return emailDeliveryResult{EmailFailures: 1}, nil
}

func emailRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	delay := time.Hour
	for index := 1; index < attempts && delay < 24*time.Hour; index++ {
		delay *= 2
	}
	if delay > 24*time.Hour {
		return 24 * time.Hour
	}
	return delay
}

func expiryEmailMessage(user *models.User, plan *models.Plan, kind string, subscriptionID uint) emailservices.Message {
	name := "FastImg"
	if user != nil && strings.TrimSpace(user.Name) != "" {
		name = user.Name
	}
	planName := "会员套餐"
	if plan != nil && strings.TrimSpace(plan.Name) != "" {
		planName = plan.Name
	}
	if user != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(user.Locale)), "zh") {
		if kind == "expired" {
			return emailservices.Message{
				To: user.Email, Subject: "FastImg 会员套餐已到期",
				Text: fmt.Sprintf("%s，您的 %s 已到期，账户已自动切换到免费套餐。图片和历史链接不会被删除。请登录 FastImg 续费或清理超出免费额度的空间。", name, planName),
				HTML: fmt.Sprintf("<p>%s，您的 <strong>%s</strong> 已到期，账户已自动切换到免费套餐。</p><p>图片和历史链接不会被删除。请登录 FastImg 续费或清理超出免费额度的空间。</p><p>订阅编号：%d</p>", html.EscapeString(name), html.EscapeString(planName), subscriptionID),
			}
		}
		return emailservices.Message{
			To: user.Email, Subject: "FastImg 会员套餐即将到期",
			Text: fmt.Sprintf("%s，您的 %s 即将到期（提醒：%s）。请及时续费，避免上传和 API 额度恢复为免费套餐。", name, planName, kind),
			HTML: fmt.Sprintf("<p>%s，您的 <strong>%s</strong> 即将到期。</p><p>提醒时间：%s。请及时续费，避免上传和 API 额度恢复为免费套餐。</p>", html.EscapeString(name), html.EscapeString(planName), html.EscapeString(kind)),
		}
	}
	if kind == "expired" {
		return emailservices.Message{
			To: user.Email, Subject: "Your FastImg plan has expired",
			Text: fmt.Sprintf("%s, your %s plan has expired and your account has been moved to the Free plan. Your media and existing links were not deleted. Please renew or reduce storage usage if you are over the Free limit.", name, planName),
			HTML: fmt.Sprintf("<p>%s, your <strong>%s</strong> plan has expired and your account has been moved to the Free plan.</p><p>Your media and existing links were not deleted.</p><p>Subscription: %d</p>", html.EscapeString(name), html.EscapeString(planName), subscriptionID),
		}
	}
	return emailservices.Message{
		To: user.Email, Subject: "Your FastImg plan will expire soon",
		Text: fmt.Sprintf("%s, your %s plan will expire soon (%s reminder). Renew in time to keep your paid upload and API limits.", name, planName, kind),
		HTML: fmt.Sprintf("<p>%s, your <strong>%s</strong> plan will expire soon.</p><p>Reminder: %s. Renew in time to keep your paid upload and API limits.</p>", html.EscapeString(name), html.EscapeString(planName), html.EscapeString(kind)),
	}
}

func (s *SubscriptionLifecycleService) publishInAppNotification(user *models.User, plan *models.Plan, kind string, endsAt *time.Time) {
	if user == nil {
		return
	}
	planName := "会员套餐"
	if plan != nil && strings.TrimSpace(plan.Name) != "" {
		planName = plan.Name
	}
	title := "会员套餐即将到期"
	body := fmt.Sprintf("%s 将在 %s 到期，请及时续费。", planName, formatExpiryTime(endsAt))
	if kind == "expired" {
		title = "会员套餐已到期"
		body = fmt.Sprintf("%s 已到期，账户已自动切换到免费套餐；图片和历史链接不会被删除。", planName)
	}
	s.notify.PublishBestEffort(user.ID, notifications.NotificationInput{
		Type:  "subscription." + kind,
		Title: title,
		Body:  body,
		URL:   "/plans",
	})
}

func formatExpiryTime(value *time.Time) string {
	if value == nil {
		return "当前周期结束"
	}
	return value.UTC().Format("2006-01-02 15:04 UTC")
}
