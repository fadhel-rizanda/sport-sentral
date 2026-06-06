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
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    PHASE 2: COMPETITION                      │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Competition  │  │  Regulator   │  │    Scout     │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  │              │  │              │  │              │      │
│  │ • Bracket    │  │ • Sport cfg  │  │ • Leaderboard      │
│  │ • Match      │  │ • Approval   │  │ • Watchlist  │      │
│  │ • Stats      │  │ • Tier       │  │ • Profiles   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   PHASE 3: COMMUNITY                         │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Event     │  │  Community   │  │    Feed      │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  │              │  │              │  │              │      │
│  │ • Pickup     │  │ • Groups     │  │ • Global     │      │
│  │ • Gathering  │  │ • Posts      │  │ • Timeline   │      │
│  │ • Entry fees │  │ • Membership │  │ • Aggregation      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   PHASE 4-5: GROWTH                          │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Payment    │  │  Attachment  │  │ Notification │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   SHARED INFRASTRUCTURE                      │
│                                                               │
│  API Gateway  │  Meta Service  │  Sport Service  │  Logger  │
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

---

## 🎓 **ACADEMY SERVICE IN DETAIL**

### **What is Academy-Service?**

Academy-Service manages the **training organization layer** of the platform:
- Academy organizational data (name, location, contact, sport)
- Athlete enrollment flow (athlete applies → admin approves)
- Roster management (team composition per competition)
- Athlete-to-Competition submission (Academy submits roster to competitions)

### **Key Relationships**

```
┌─────────────────────────────────────────────────────┐
│                    Academy                           │
│                                                      │
│  Sport: Basketball (tenant-bound)                   │
│  Admin: academy_admin user                          │
│  Location: Jakarta, GOR Senayan                     │
│                                                      │
│  └─ Enrollments (Athletes in this academy)          │
│     ├─ Athlete 1 (enrolled, status: ACTIVE)        │
│     ├─ Athlete 2 (enrolled, status: ACTIVE)        │
│     └─ Athlete 3 (soft-deleted, stats intact)      │
│                                                      │
│  └─ Roster (for specific competition)               │
│     ├─ Roster A (competition X, 5 players)         │
│     └─ Roster B (competition Y, 3 players)         │
│                                                      │
│  └─ Participants (submitted to competitions)        │
│     ├─ Participant A (team, 5 athletes)            │
│     └─ Participant B (individual, 1 athlete)       │
└─────────────────────────────────────────────────────┘
```

### **Core Entities in Academy-Service**

#### **1. Academy**
- One academy = one sport (tenant-bound)
- Unique per sport + region
- Managed by academy_admin role
- Requires platform_admin approval on creation

**Fields:**
```go
type Academy struct {
    ID             uuid.UUID  // Primary key
    Name           string     // "Jakarta Hoops Academy"
    Description    string     // About the academy
    Sport          string     // Which sport (basketball, volleyball, etc)
    AdminID        uuid.UUID  // FK to academy_admin user
    City           string     // Jakarta, Surabaya, etc
    Address        string     // Physical location
    PhoneNumber    string     // Contact
    Email          string     // Contact
    Founded        time.Time  // When academy was established
    ImageURL       string     // Academy logo/image
    Status         string     // PENDING, APPROVED, REJECTED
    ApprovedByID   *uuid.UUID // FK to platform_admin
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt // Soft delete
}
```

#### **2. Enrollment** (Athlete ↔ Academy)
- Many-to-many relationship with soft-delete safety
- Status: ACTIVE, SUSPENDED, LEFT
- When athlete leaves: soft-delete enrollment, stats remain via Sport-Athlete (separate service)

**Fields:**
```go
type Enrollment struct {
    ID             uuid.UUID  // Primary key
    AcademyID      uuid.UUID  // FK Academy
    AthleteID      uuid.UUID  // FK User (athlete role)
    Status         string     // PENDING, ACTIVE, SUSPENDED, LEFT
    JoinedAt       time.Time
    ApprovedByID   *uuid.UUID // Academy admin who approved
    ApprovedAt     *time.Time
    LeftAt         *time.Time // When athlete left
    DeletedAt      gorm.DeletedAt // Soft delete (keeps stats)
}
```

#### **3. Roster** (For Competition Submission)
- Collection of athletes from an academy for a specific competition
- Type: team | individual (based on sport config)
- Submitted to competition-service as "Participant"

**Fields:**
```go
type Roster struct {
    ID             uuid.UUID
    AcademyID      uuid.UUID
    CompetitionID  uuid.UUID  // Optional: which competition this roster is for
    Name           string     // "Jakarta Hoops Basketball Team"
    RosterType     string     // TEAM | INDIVIDUAL
    MaxSize        int        // From sport config
    CreatedAt      time.Time
    
    // Relationships
    Athletes []RosterMember // Flexible roster members
}

type RosterMember struct {
    ID             uuid.UUID
    RosterID       uuid.UUID
    AthleteID      uuid.UUID
    JerseyNumber   int        // Optional
    Position       string     // Optional: "Guard", "Forward", etc
    AddedAt        time.Time
}
```

### **Academy-Service Responsibilities**

```
┌─────────────────────────────────────────┐
│    ACADEMY-SERVICE RESPONSIBILITIES     │
├─────────────────────────────────────────┤
│                                         │
│  1. ACADEMY MANAGEMENT                  │
│     • Create academy (admin approval)   │
│     • Update academy info               │
│     • List academies per sport          │
│     • Get academy details               │
│                                         │
│  2. ENROLLMENT FLOW                     │
│     • Athlete applies to academy        │
│     • Academy admin reviews             │
│     • Academy admin approves/rejects    │
│     • Track enrollment status           │
│     • Athlete can leave academy         │
│                                         │
│  3. ROSTER MANAGEMENT                   │
│     • Create roster (team + athletes)   │
│     • Add/remove athlete from roster    │
│     • Validate roster (min/max size)    │
│     • Submit roster to competition      │
│                                         │
│  4. DATA QUERIES                        │
│     • List enrollments per academy      │
│     • List enrollments per athlete      │
│     • Get athlete enrollment status     │
│     • Export roster (for print/submit)  │
│                                         │
│  5. INTEGRATION POINTS                  │
│     → Identity-Service: user info       │
│     → Meta-Service: status/tag data     │
│     → Sport-Service: sport config       │
│     → Competition-Service: roster submit│
│                                         │
└─────────────────────────────────────────┘
```

---

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

Want me to continue with the full implementation code?