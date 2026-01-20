# AI Productivity Coach  
Scale-ready backend + ML/LLM system  
Version: Draft 1

## Overview
The AI Productivity Coach helps users improve task execution consistency by tracking behavior, detecting procrastination patterns, providing intelligent scheduling suggestions, and generating humorous coaching nudges. The system aims to increase follow-through on tasks and support long-term behavioral improvement.

## Problem Statement
Most productivity tools track tasks but do not actively coach users or adapt to their behavioral patterns. Users procrastinate, snooze tasks, and misjudge workload without feedback. Introducing adaptive coaching and suggestions can improve adherence and reduce drop-off.

## Target Users
Primary:
- Knowledge workers managing mixed personal and professional tasks
- Students with deadline-driven workloads
- Self-directed workers, founders, creators, freelancers

Secondary:
- Habit and streak-based productivity users
- Productivity enthusiasts

## Objectives
- Increase task completion rates
- Reduce procrastination through modeling and nudges
- Personalize scheduling and recommendations
- Introduce humor and personality without harming usability
- Support long-term engagement and improvement

## Key Features (User-Facing)
- Task CRUD: create, edit, snooze, complete, delete
- Scheduling recommendations
- Behavioral tracking and procrastination detection
- Coaching nudges via notifications
- Humor modes and coaching personas
- User preferences for notifications, humor, work hours

## Key Features (System)
- Behavioral event logging
- ML-based scoring and clustering
- LLM-based messaging generator
- Recommendation and scheduling engine
- Feature store for ML
- Event-driven asynchronous pipeline
- Analytics storage and retrieval
- Gateway for gRPC-based clients

## Non-Functional Requirements
**Performance**
- Task CRUD p95 < 200ms
- Notification dispatch within target window
- On-demand nudges 1–3s latency
- Batch workloads isolated from interactive traffic

**Scalability**
- Horizontal scaling for behavior, coaching, and notification services
- Queue-based backpressure and retries
- Separation of online vs offline workloads

**Reliability**
- Durable event logs
- Zero data loss for task metadata
- Degraded mode if LLM unavailable

**Security**
- Encrypted in transit and at rest
- Token-based authentication
- Controlled access to behavioral analytics

**Compliance & Privacy**
- Data export and deletion
- Behavioral profiling transparency
- Region-specific data handling roadmap

**Extensibility**
- Multi-client support (web, mobile, future devices)
- Persona modes for coaching
- Additional ML models for streaks and habits

## System Constraints
- LLM must not block core task CRUD
- ML must run batch + online scoring
- Behavioral data must be stored efficiently for long-term analysis
- Cost and inference must be controlled

## Behavioral Modeling Requirements
**Signals**
- Completion timing vs assigned
- Snooze frequency
- Edit frequency
- Deadline tension
- Inactivity windows
- Personal scheduling rhythm
- Category clustering

**Outputs**
- Procrastination score
- Scheduling suggestions
- Nudge triggers
- Persona adjustments

## LLM Coaching Requirements
**Modes**
- Encouragement
- Roast
- Neutral
- Minimalist

**Context Inputs**
- Tasks
- Scores
- Time of day
- Deadlines
- Preferences

**Fallback Behavior**
- Template nudges

**Error Handling**
- Timeouts and retries via queue

## Success Metrics
**Primary**
- Increase in completed tasks
- Reduction in overdue tasks
- Reduction in snooze rate
- Retention (weekly active usage)

**Secondary**
- User satisfaction
- Drop-off after onboarding
- Nudge engagement

**ML/LLM Metrics**
- Nudge usefulness rating
- Persona engagement
- Notification misfire rate

## Anti-Goals (v1)
- No collaboration or shared tasks
- No enterprise project management features
- No heavy habit-only mode
- No financial or medical advice
- No enterprise authentication or SSO

## Assumptions
- Users treat tool as personal assistant
- Behavioral tracking is opt-in
- Humor remains non-offensive
- Minimal setup preferences

## Future Extensions (Optional)
- Wearable integrations
- Calendar sync
- Workday inference
- Team/manager coaching
- On-device inference

## ER Model Fields and Justifications

## User

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Global unique identity across services |
| email | Text | Unique | Authentication and communication |
| created_at | Timestamp | No | Account lifecycle analytics |
| timezone | Text | No | Required for scheduling alignment |
| work_hours | JSON | No | Personalized scheduling window |
| deleted_at | Timestamp | No | Soft delete for GDPR compliance |

---

## Task

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique task identity |
| user_id | UUID | FK→User.id | Ownership |
| title | Text | No | Minimal summary |
| description | Text | No | Context for LLM coaching |
| status | Enum(pending,completed,snoozed,deleted) | No | Controls scheduling and analytics |
| due_at | Timestamp | No | Target date/time for scheduling |
| category | Text | No | Clustering and personalization |
| priority | Smallint | No | Manual scheduling hint |
| estimated_minutes | Integer | No | Scheduling and duration modeling |
| created_at | Timestamp | No | Lifecycle tracking |
| updated_at | Timestamp | No | Sync and analytics |
| deleted_at | Timestamp | No | Soft delete |

---

## TaskEvent

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique event |
| task_id | UUID | FK→Task.id | Relates to task |
| user_id | UUID | FK→User.id | Behavior analytics |
| event_type | Enum(create,modify,snooze,complete) | No | Event classification |
| event_payload | JSON | No | Flexible event metadata |
| created_at | Timestamp | No | Time-series analytics |

---

## ScheduleRecommendation

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique suggestion |
| user_id | UUID | FK→User.id | Target user |
| task_id | UUID | FK→Task.id | Target task |
| recommended_due | Timestamp | No | Suggested scheduling |
| confidence_score | Float | No | ML explainability |
| created_at | Timestamp | No | Traceability |
| accepted | Boolean | No | Feedback loop |
| rejected | Boolean | No | Model tuning |

---

## CoachingNudge

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique nudge instance |
| user_id | UUID | FK→User.id | Target user |
| task_id | UUID | FK→Task.id | Optional task linkage |
| mode | Enum(roast,encouragement,neutral,minimalist) | No | Persona control |
| generated_text | Text | No | LLM or template content |
| source | Enum(LLM,Template) | No | Cost and fallback path |
| created_at | Timestamp | No | Analytics |
| delivered | Boolean | No | Delivery tracking |
| delivery_channel | Enum(push,email,sms) | No | Routing |
| sentiment_score | Float | No | Tone analysis |

---

## UserPreference

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique preference record |
| user_id | UUID | FK→User.id | Ownership |
| humor_intensity | Smallint | No | Personalization |
| notification_window | JSON | No | Delivery control |
| nudges_enabled | Boolean | No | Opt-out flag |
| created_at | Timestamp | No | Versioning |
| updated_at | Timestamp | No | Sync correctness |

---

## Notification

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique notification event |
| user_id | UUID | FK→User.id | Routing |
| nudge_id | UUID | FK→CoachingNudge.id | Traceability |
| status | Enum(queued,sent,failed) | No | Observability |
| channel | Enum(push,email,sms) | No | Multichannel delivery |
| created_at | Timestamp | No | Auditing |
| delivered_at | Timestamp | No | Delivery latency metrics |

---

## FeatureSnapshot

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique snapshot |
| user_id | UUID | FK→User.id | Scope for ML |
| procrastination_score | Float | No | Core ML output |
| snooze_rate | Float | No | Derived behavioral feature |
| completion_latency_avg | Float | No | Time-based signal |
| burstiness_index | Float | No | Work pattern characterization |
| mode_cluster | Integer | No | Clustering label |
| created_at | Timestamp | No | Time-series trend tracking |

---

## PersonaProfile

| Field | Type | PK | Reason |
|---|---|---|---|
| id | UUID | Yes | Unique persona record |
| user_id | UUID | FK→User.id | Ownership |
| persona_type | Enum(roast,chill,coach,neutral) | No | UX persona |
| adaptation_enabled | Boolean | No | ML tuning |
| created_at | Timestamp | No | Lifecycle |
| updated_at | Timestamp | No | Dynamic persona control |