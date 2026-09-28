package defaults

import "encoding/json"

// Page describes the content that is created for a new FastImg installation.
// The defaults are intentionally English so a fresh public site is usable for
// an international audience. Administrators can replace them from the page
// editor after installation.
type Page struct {
	Slug           string
	Title          string
	Excerpt        string
	SEOTitle       string
	SEODescription string
	OldTitle       string
	OldExcerpt     string
	Content        string
}

var pages = []Page{
	{
		Slug:           "privacy",
		Title:          "Privacy Policy",
		Excerpt:        "How FastImg collects, uses, stores, and protects account, media, and access data.",
		SEOTitle:       "Privacy Policy | FastImg",
		SEODescription: "Learn how FastImg handles account information, uploaded media, access logs, cookies, and service providers.",
		OldTitle:       "隐私政策",
		OldExcerpt:     "FastImg 如何处理账户、媒体和访问数据。",
		Content: document([]node{
			heading("Privacy Policy", 1),
			paragraph("Effective date: September 28, 2026"),
			heading("Overview", 2),
			paragraph("FastImg provides media hosting, stable links, image delivery, API access, and media management for developers, site owners, and content creators. This Privacy Policy explains what information we process, why we process it, and the choices available to you when you use the service."),
			heading("Information we collect", 2),
			paragraph("We collect the information needed to operate and secure your account, including your email address, authentication records, plan and billing status, support requests, and security events. When you upload media, we process the file, filename, MIME type, size, checksum, visibility setting, storage location, and related metadata."),
			paragraph("We also collect operational data such as request timestamps, IP address, user agent, API token usage, delivery results, bandwidth estimates, and error records. We use reasonable data minimization: logs are retained only for security, quota accounting, troubleshooting, compliance, and service improvement."),
			heading("How we use information", 2),
			paragraph("We use information to authenticate users, store and deliver media, enforce plan limits, generate links, process payments, send service notifications, prevent abuse, investigate reports, answer support requests, and maintain the availability and security of FastImg. We do not sell personal information."),
			heading("Media visibility and sharing", 2),
			paragraph("Uploads are public by default so that their generated links can be embedded in websites, forums, and applications. You can change a media item to private when supported by your plan. Private media requires an authorized request or a valid signed URL; having an identifier alone does not grant access."),
			heading("Service providers", 2),
			paragraph("FastImg may use infrastructure, object storage, email, analytics, and payment providers to deliver the service. These providers receive only the information required for their function and must protect it under their applicable terms and privacy commitments."),
			heading("Retention and deletion", 2),
			paragraph("You may delete your media and account data through the available controls, subject to a short recovery period, security records, legal requirements, and backup retention. Deleted media is removed from active delivery and is permanently cleaned up according to the configured retention policy."),
			heading("Security", 2),
			paragraph("We use access controls, encrypted secrets, signed delivery URLs, audit records, rate limits, and isolated administrative permissions to protect the service. No online service can guarantee absolute security, so please use a unique password, protect API tokens, and report suspected abuse promptly."),
			heading("Your choices and contact", 2),
			paragraph("You can update your account details, revoke API tokens, change media visibility, manage plan settings, and request account assistance from the service interface. For privacy questions or data requests, contact the site administrator through the configured support address."),
		}),
	},
	{
		Slug:           "terms",
		Title:          "Terms of Use",
		Excerpt:        "The rules for using FastImg responsibly, safely, and within your selected plan.",
		SEOTitle:       "Terms of Use | FastImg",
		SEODescription: "Read the FastImg terms for accounts, uploads, acceptable use, plans, payments, moderation, and service availability.",
		OldTitle:       "使用条款",
		OldExcerpt:     "使用 FastImg 服务时需要遵守的基本规则。",
		Content: document([]node{
			heading("Terms of Use", 1),
			paragraph("Effective date: September 28, 2026"),
			heading("Acceptance", 2),
			paragraph("By creating an account, uploading media, or using a FastImg link, API, plan, or public page, you agree to these Terms of Use and the Privacy Policy. If you use FastImg on behalf of an organization, you confirm that you have authority to accept these terms for that organization."),
			heading("Your account", 2),
			paragraph("You are responsible for providing accurate registration information, protecting your password and API tokens, and all activity performed through your account. Do not share credentials or use another person’s account without permission. Notify the site administrator promptly if you suspect unauthorized access."),
			heading("Acceptable uploads", 2),
			paragraph("You may upload only content that you own or are authorized to host and distribute. You remain responsible for the content, metadata, links, and claims associated with your media. You grant FastImg the limited permission needed to store, process, cache, display, and deliver your media as part of the service."),
			heading("Prohibited use", 2),
			paragraph("You must not use FastImg to host or distribute unlawful, infringing, deceptive, abusive, hateful, sexually exploitative, malicious, or privacy-invasive material. You must not upload malware, conduct attacks, bypass quotas or access controls, probe private endpoints, send spam, scrape the service at an unreasonable rate, or interfere with other users."),
			heading("Public media and moderation", 2),
			paragraph("Public media may be viewed and embedded by anyone who receives a valid link. FastImg may hide, restrict, remove, or investigate media after a report, security signal, legal request, or policy violation. A report is not a finding of wrongdoing, and enforcement decisions may consider the available evidence and applicable law."),
			heading("Plans, limits, and payments", 2),
			paragraph("Free and paid plans provide the storage, file-size, bandwidth, API, processing, and feature limits shown at the time of selection. Paid access begins when the selected payment provider confirms the transaction. Expired or canceled plans may be downgraded to the configured default plan, while existing data is handled according to the retention and quota rules displayed in the service."),
			heading("Availability and changes", 2),
			paragraph("FastImg may perform maintenance, change limits, add or remove integrations, or temporarily suspend delivery to protect the service. We will make reasonable efforts to communicate material changes. These terms may be updated from time to time; continued use after an update means that you accept the revised terms."),
			heading("Termination and contact", 2),
			paragraph("You may stop using FastImg at any time. We may suspend or terminate accounts that violate these terms, create security or legal risk, or remain unpaid. If you have a question about an account action, payment, or content report, contact the site administrator through the configured support channel."),
		}),
	},
	{
		Slug:           "about",
		Title:          "About FastImg",
		Excerpt:        "FastImg is a membership-based media hosting platform for developers, site owners, and content creators.",
		SEOTitle:       "About FastImg",
		SEODescription: "Discover FastImg, a practical media hosting platform with stable links, API uploads, access control, and asset management.",
		OldTitle:       "关于我们",
		OldExcerpt:     "FastImg 面向开发者、站长和内容创作者。",
		Content: document([]node{
			heading("About FastImg", 1),
			heading("A focused home for your media", 2),
			paragraph("FastImg is a membership-based media hosting platform built for developers, site owners, and content creators who need dependable links and straightforward asset management. It combines a clean upload experience with the controls needed to run media reliably across websites, forums, applications, and documentation."),
			heading("What you can do", 2),
			paragraph("Upload images from the web interface or API, copy ready-to-use links in several formats, organize assets with folders and albums, manage public and private visibility, create signed links, review usage, and control API tokens from one account."),
			paragraph("Administrators can manage members, plans, quotas, advertisements, storage connections, reports, content pages, footer navigation, payments, access logs, and platform settings. Member features and administrative operations are intentionally separated so each user can manage their own assets without entering the management console."),
			heading("Built for practical operation", 2),
			paragraph("FastImg is designed around stable URLs, explicit ownership, plan-aware limits, auditable administrative actions, safe rich-text content, and provider adapters that can be enabled as the deployment grows. The platform can start with local storage and later use compatible object storage or CDN infrastructure."),
			heading("Our approach", 2),
			paragraph("We prefer clear defaults, understandable controls, and honest operational data over hidden complexity. Usage numbers are derived from recorded media and access events, while payment and storage integrations expose only the capabilities that are actually configured."),
			heading("Get started", 2),
			paragraph("Create an account to upload your first image, explore the available plans, or browse the public discovery and friend-link pages without signing in. For platform questions, partnership requests, or abuse reports, use the contact channel configured by the site administrator."),
		}),
	},
}

type node map[string]any

func DefaultPages() []Page {
	result := make([]Page, len(pages))
	copy(result, pages)
	return result
}

func document(nodes []node) string {
	content := make([]map[string]any, 0, len(nodes))
	for _, item := range nodes {
		content = append(content, map[string]any(item))
	}
	encoded, _ := json.Marshal(map[string]any{"type": "doc", "content": content})
	return string(encoded)
}

func heading(text string, level int) node {
	return node{"type": "heading", "attrs": map[string]any{"level": level}, "content": []any{map[string]any{"type": "text", "text": text}}}
}

func paragraph(text string) node {
	return node{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": text}}}
}
