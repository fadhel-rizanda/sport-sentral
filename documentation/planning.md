# Sport-Sentral: Complete Project Planning & Documentation

Baik, saya sudah baca planning document-nya. Mari saya breakdown dalam format markdown yang lebih terstruktur, khususnya fokus ke **academy-service** yang mau kita buat sekarang.

---

## 📄 **SPORT-SENTRAL PROJECT OVERVIEW**

**Sport-Sentral** adalah centralized multi-sport ecosystem platform yang menghubungkan:
- 🏃 Athletes
- 🏫 Academies
- 🏀 Court Owners
- 🔍 Scouts
- 👮 Regulators (Sport Bodies)
- 🎯 Event Organizers
- 👥 Community Members

**Tagline:** Data legitimacy, role flexibility, ecosystem centralization

---

## 🏗️ **ARCHITECTURE OVERVIEW**

### **Microservices Map**

```
┌─────────────────────────────────────────────────────────────┐
│                      PHASE 1: FOUNDATION                    │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Identity    │  │   Academy    │  │    Court     │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  │              │  │              │  │              │      │
│  │ • Auth       │  │ • Academies  │  │ • Inventory  │      │
│  │ • Users      │  │ • Enrollment │  │ • Booking    │      │
│  │ • RBAC       │  │ • Roster     │  │ • Schedule   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│                    PHASE 2: COMPETITION                    │
│                                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Competition  │  │    Sport     │  │    Scout     │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  │              │  │              │  │              │      │
│  │ • Bracket    │  │ • Sport cfg  │  │ • Leaderboard       │
│  │ • Match      │  │ • Approval   │  │ • Watchlist  │      │
│  │ • Stats      │  │ • Tier       │  │ • Profiles   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│                   PHASE 3: COMMUNITY                       │
│                                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Event     │  │  Community   │  │    Feed      │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  │              │  │              │  │              │      │
│  │ • Pickup     │  │ • Groups     │  │ • Global     │      │
│  │ • Gathering  │  │ • Posts      │  │ • Timeline   │      │
│  │ • Entry fees │  │ • Membership │  │ • Aggregation       │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│                   PHASE 4-5: GROWTH                        │
│                                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Payment    │  │  Attachment  │  │ Notification │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐                                          │
│  │  Audit Log   │                                          │
│  │   Service    │                                          │
│  └──────────────┘                                          │
└────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   SHARED INFRASTRUCTURE                     │
│                                                             │
│       API Gateway  │  Meta Service  │  Logger  │            │
└─────────────────────────────────────────────────────────────┘
```

### **Current Status (Done ✅)**
- ✅ Identity Service (Auth, Users, RBAC)
- ✅ Meta Service (Status, Tags)
- ✅ API Gateway (REST ↔ gRPC)
- ✅ Telemetry (Jaeger, Prometheus, Grafana)

### **Next Priority: Academy Service (This Phase)**
- 🔄 Academy Service (Tenant management, enrollment, roster)

---

## 👥 **ROLE SYSTEM**

### **Role Matrix**

| Role | Self-Register | Verification | General Access | Notes |
|------|---------------|--------------|----------------|-------|
| **athlete** | ✅ Yes | ❌ No | ✅ Full | Default role, instant active |
| **scout** | ✅ Yes | ❌ No | ✅ Limited (freemium) | Talent finder |
| **court_owner** | ✅ Yes | ✅ Admin approval per account + per court | ✅ Full | Manages physical facilities |
| **academy_admin** | ✅ Yes | ✅ Admin approval per account + per academy | ✅ Full | Manages academy tenant |
| **regulator** | ❌ No | ✅ Admin-assigned only | ✅ Full | Official sport body (PERBASI, FIBA) |
| **platform_admin** | ❌ No | ✅ Internal only | ✅ Full | Super admin |

### **Key Rules**
- 🔒 **Permanent Role**: Once chosen at registration, role CANNOT be changed
- 👤 **Multi-Sport Identity**: Athlete can participate in multiple academies (per sport)
- 🏢 **Academy Context**: Stats are tied to academy + sport, not just user
- 🧹 **Soft Delete**: When athlete leaves academy, enrollment is soft-deleted but stats remain
- 🔑 **RBAC-Driven Membership & Authorization**: 
  - Domain services (`scout-service`, `academy-service`, `venue-service`) do **NOT** maintain hardcoded tier fields (`tier`, `is_premium`).
  - Feature gating and membership privileges are driven by **Role-Permissions (RBAC)**.
  - When payment/subscription feature is added in Phase 4/5, `payment-service` will trigger updates in `identity-service` (RBAC) to grant or revoke user roles/permissions upon payment events.

---

## 🔐 **MEMBERSHIP & AUTHORIZATION ARCHITECTURE**

```
┌─────────────────┐       ┌─────────────────────────────────┐       ┌────────────────────────┐
│  User Purchase  │ ───►  │  Payment Service (Phase 4/5)    │ ───►  │  Identity Service /    │
│  (Subscription) │       │  • Handles Payment Webhooks     │       │  RBAC                  │
└─────────────────┘       │  • Manages Expirations & Crons  │       │  • Grants/Revokes      │
                          └─────────────────────────────────┘       │    Roles & Permissions │
                                                                    └───────────┬────────────┘
                                                                                │
                                                                                │ Enforces Permissions
                                                                                ▼
                                                                    ┌────────────────────────┐
                                                                    │ Domain Services        │
                                                                    │ (Scout, Academy, Venue)│
                                                                    │ • No custom tier fields│
                                                                    │ • Standard RBAC checks │
                                                                    └────────────────────────┘
```

## 📜 **BUSINESS AUDIT & ACTIVITY LOG ARCHITECTURE**

```
┌─────────────────┐      NATS / Kafka       ┌────────────────────────┐
│ Domain Services │ ─────────────────────┐  │ Audit / Activity       │
│ (Scout, Academy,│  Publishes Domain    │  │ Log Service            │
│  Venue, etc.)   │  Events Async        ├─►│ • Append-only Store    │
└─────────────────┘                      │  │ • Unified Activity Feed│
                                         │  │ • Zero User Request    │
                                         │  │   Latency Impact       │
                                         └─►└────────────────────────┘
```

### **Core Principles**
- 🛡️ **Separation of Concerns**: Business process level logs (e.g. *"Scout A added Athlete B to watchlist"*, *"Academy C enrolled Athlete D"*) are isolated from domain service databases into a dedicated **Audit Log Service** (Phase 4-5).
- ⚡ **Async Event-Driven Logging**: Domain services publish domain events asynchronously to the message broker (NATS/Kafka). The Audit Log Service consumes events without adding latency to primary user HTTP/gRPC requests.
- 🔒 **Immutable & Tamper-Proof**: Audit trails are append-only. Domain services cannot mutate or delete historical activity logs.
- 🔍 **Unified Global Activity Feed**: Enables administrators and users to query a consolidated timeline of activity across the entire platform.

## 📊 **ACADEMY-SERVICE DATABASE SCHEMA**

```sql
-- Academies Table
CREATE TABLE academies (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    sport VARCHAR(50) NOT NULL,  -- basketball, volleyball, etc
    admin_id UUID NOT NULL,       -- FK to identity.users
    city VARCHAR(100),
    address TEXT,
    phone_number VARCHAR(20),
    email VARCHAR(255),
    founded_at TIMESTAMP,
    image_url TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',  -- PENDING, APPROVED, REJECTED
    approved_by_id UUID,          -- FK to identity.users (platform_admin)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_admin FOREIGN KEY (admin_id) REFERENCES identity.users(id),
    CONSTRAINT fk_approved_by FOREIGN KEY (approved_by_id) REFERENCES identity.users(id),
    INDEX idx_sport (sport),
    INDEX idx_city (city),
    INDEX idx_status (status),
    INDEX idx_admin (admin_id)
);

-- Enrollments Table (Athlete ↔ Academy)
CREATE TABLE enrollments (
    id UUID PRIMARY KEY,
    academy_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',  -- PENDING, ACTIVE, SUSPENDED, LEFT
    joined_at TIMESTAMP NOT NULL,
    approved_by_id UUID,          -- FK to identity.users (academy_admin)
    approved_at TIMESTAMP,
    left_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_academy FOREIGN KEY (academy_id) REFERENCES academies(id) ON DELETE CASCADE,
    CONSTRAINT fk_athlete FOREIGN KEY (athlete_id) REFERENCES identity.users(id),
    CONSTRAINT fk_approved_by FOREIGN KEY (approved_by_id) REFERENCES identity.users(id),
    UNIQUE KEY uq_academy_athlete (academy_id, athlete_id),
    INDEX idx_academy (academy_id),
    INDEX idx_athlete (athlete_id),
    INDEX idx_status (status)
);

-- Rosters Table (for competition submission)
CREATE TABLE rosters (
    id UUID PRIMARY KEY,
    academy_id UUID NOT NULL,
    competition_id UUID,          -- Optional: which competition
    name VARCHAR(255) NOT NULL,
    roster_type VARCHAR(20),      -- TEAM, INDIVIDUAL
    max_size INT DEFAULT 12,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_academy FOREIGN KEY (academy_id) REFERENCES academies(id) ON DELETE CASCADE,
    INDEX idx_academy (academy_id),
    INDEX idx_competition (competition_id)
);

-- Roster Members Table
CREATE TABLE roster_members (
    id UUID PRIMARY KEY,
    roster_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    jersey_number INT,
    position VARCHAR(50),
    added_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT fk_roster FOREIGN KEY (roster_id) REFERENCES rosters(id) ON DELETE CASCADE,
    CONSTRAINT fk_athlete FOREIGN KEY (athlete_id) REFERENCES identity.users(id),
    UNIQUE KEY uq_roster_athlete (roster_id, athlete_id),
    INDEX idx_roster (roster_id),
    INDEX idx_athlete (athlete_id)
);
```

---

## 🔌 **API ENDPOINTS (gRPC)**

### **Academy Management**

```protobuf
service AcademyService {
    // Academy CRUD
    rpc GetAcademy(GetAcademyRequest) returns (Academy);
    rpc ListAcademies(ListAcademiesRequest) returns (ListAcademiesResponse);
    rpc CreateAcademy(CreateAcademyRequest) returns (Academy);
    rpc UpdateAcademy(UpdateAcademyRequest) returns (Academy);
    rpc DeleteAcademy(DeleteAcademyRequest) returns (google.protobuf.Empty);
    
    // Enrollment Management
    rpc ApplyToAcademy(ApplyRequest) returns (Enrollment);
    rpc ApproveEnrollment(ApproveEnrollmentRequest) returns (Enrollment);
    rpc RejectEnrollment(RejectEnrollmentRequest) returns (google.protobuf.Empty);
    rpc LeaveAcademy(LeaveAcademyRequest) returns (google.protobuf.Empty);
    rpc GetEnrollment(GetEnrollmentRequest) returns (Enrollment);
    rpc ListEnrollments(ListEnrollmentsRequest) returns (ListEnrollmentsResponse);
    
    // Roster Management
    rpc CreateRoster(CreateRosterRequest) returns (Roster);
    rpc AddAthletesToRoster(AddAthletesRequest) returns (Roster);
    rpc RemoveAthleteFromRoster(RemoveAthleteRequest) returns (Roster);
    rpc SubmitRosterToCompetition(SubmitRosterRequest) returns (Participant);
    rpc GetRoster(GetRosterRequest) returns (Roster);
}
```

---

## 🔄 **ATHLETE ENROLLMENT FLOW**

```
1. ATHLETE APPLIES
   athlete user
   └─ calls: ApplyToAcademy(academy_id)
      └─ status: PENDING
      
2. ACADEMY ADMIN REVIEWS
   academy_admin logs in → sees pending enrollments
   └─ calls: ApproveEnrollment(enrollment_id)
      OR RejectEnrollment(enrollment_id)
      
3. ATHLETE ACTIVE IN ACADEMY
   athlete.enrollment.status = ACTIVE
   └─ Athlete can now be added to rosters
   └─ Stats will be tracked under this academy
   
4. ATHLETE LEAVES (Soft Delete)
   athlete calls: LeaveAcademy(academy_id)
   └─ enrollment soft-deleted (deleted_at set)
   └─ Stats remain intact (for historical records)
   └─ Athlete can apply to different academy (same sport) or different sport
```

---

## 📋 **IMPLEMENTATION ROADMAP**

### **Phase 1: Core Structure (This Sprint)**
- ✅ Proto files (Academy, Enrollment, Roster messages)
- ✅ Database entities & migrations
- ✅ Repository layer
- ✅ Use case layer with telemetry
- ✅ gRPC handlers
- ✅ Config & main setup

### **Phase 2: Integration (Next Sprint)**
- Integration with identity-service (user verification)
- Integration with meta-service (status tracking)
- Integration with sport-service (sport config validation)
- Event publishing (enrollment approved → notify athlete)

### **Phase 3: Enhancement (Later)**
- Bulk enrollment import (CSV upload)
- Roster export functionality
- Academy search & discovery
- Analytics dashboard

---

## 🚀 **NEXT STEPS**

Ready to implement? Here's what we'll build:

1. **Proto files** - Define all gRPC messages
2. **Database schema** - Create migrations
3. **Entity models** - Go structs
4. **Repositories** - Data access layer
5. **Use cases** - Business logic with tracing
6. **Handlers** - gRPC endpoints
7. **Config & Main** - Service setup with telemetry
8. **Docker integration** - Container setup
