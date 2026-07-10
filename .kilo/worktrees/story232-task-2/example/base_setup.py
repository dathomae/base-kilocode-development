# This file contains the basic database state upon which all the
# examples in this directory depend. This is human-oriented
# documentation, simplified to be demonstrative. LLMs query
# MCP query which answers with more-detailed JSON results. An
# exhaustive case study would have considerably more detail;
# however, the system itself is designed to process that detail
# at scale.
#
#   ns_family: sales, type=plan of record
#       ns: plan of record, type=Production
#       ns: market shock, type=Scenario
#       ns: viral story, type=Scenario
#
#   category: customer journey (ns: plan of record)
#       elem: awareness & pre-intent
#       elem: intent & early engagement
#       elem: discovery & evaluation
#       elem: direct interaction & qualification
#       elem: trial, testing & commitment
#       elem: purchase & conversion
#       elem: post-purchase & retention
#       elem: advocacy & network effects
#
#   category: customer awareness (ns: plan of record, taxonomy_parent_id: awareness & pre-intent)
#       elem: impression, exposure, or view
#       elem: read blog, watch video
#       elem: click through or open email
#       elem: social media engagement
#       elem: brand search
#
#   category: customer intent (ns: plan of record, taxonomy_parent_id: intent & early engagement)
#       elem: first web visit
#       elem: return web visit
#       elem: landing page engagement
#       elem: lead capture or newsletter subscription
#
#   category: product evaluation (ns: plan of record, taxonomy_parent_id: discovery & evaluation)
#       elem: Feature Comparison
#       elem: Pricing Page Visit
#       elem: Configurator or Calculator Use
#       elem: Documentation, FAQ Engagement, or Demo Video
#       elem: Case Study View
#
#   category: customer qualification (ns: plan of record, taxonomy_parent_id: direct interaction & qualification)
#       elem: Product Inquiry (your existing category)
#       elem: Contact Form Submission
#       elem: Chatbot Interaction
#       elem: Sales Call / Meeting
#       elem: RFP / RFQ Submission
#       elem: Lead Qualification (MQL → SQL)
#
#   category: customer commitment (ns: plan of record, taxonomy_parent_id: trial, testing, & commitment)
#       elem: Product Trial (your existing category)
#       elem: Proof of Concept (POC)
#       elem: Pilot Program
#       elem: Free Tier Activation
#       elem: Usage Milestones (activation, onboarding steps)
#
#   category: customer purchase (ns: plan of record, taxonomy_parent_id: purchase & conversion)
#       elem: Product Purchase (your existing category)
#       elem: Checkout Initiation
#       elem: Cart Abandonment
#       elem: Contract Signature
#       elem: Subscription Start
#
#   category: customer retention (ns: plan of record, taxonomy_parent_id: post-purchase & retention)
#       elem: Onboarding Completion
#       elem: First Value / Activation Event
#       elem: Product Usage
#       elem: Support Ticket
#       elem: Customer Satisfaction
#       elem: Renewal
#       elem: Upsell / Cross‑sell
#       elem: Churn / Expiry
#       elem: Refund / Cancellation
#
#   category: customer advocacy (ns: plan of record, taxonomy_parent_id: advocacy & network effects)
#       elem: Review / Rating
#       elem: Referral
#       elem: Community Participation
#       elem: Case Study Participation
#
#   import_taxonomy
#       dest=ns: market shock
#       source=ns: plan of record, name: customer journey
#   import_taxonomy
#       dest=ns: viral story
#       source=ns: plan of record, name: customer journey
#
#   ts_doc_type: sales forecast (ns: plan of record)
#       ts_doc_type_row: