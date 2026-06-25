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
│                    Academy                          │
│                                                     │
│  Sport: Basketball (tenant-bound)                   │
│  Admin: academy_admin user                          │
│  Location: Jakarta, GOR Senayan                     │
│                                                     │
│  └─ Enrollments (Athletes in this academy)          │
│     ├─ Athlete 1 (enrolled, status: ACTIVE)         │
│     ├─ Athlete 2 (enrolled, status: ACTIVE)         │
│     └─ Athlete 3 (soft-deleted, stats intact)       │
│                                                     │
│  └─ Roster (for specific competition)               │
│     ├─ Roster A (competition X, 5 players)          │
│     └─ Roster B (competition Y, 3 players)          │
│                                                     │
│  └─ Participants (submitted to competitions)        │
│     ├─ Participant A (team, 5 athletes)             │
│     └─ Participant B (individual, 1 athlete)        │
└─────────────────────────────────────────────────────┘
```

### **Core Entities in Academy-Service**

#### **1. Academy**
- One academy = one sport (tenant-bound)
- Unique per sport + region
- Managed by academy_admin role
- Requires platform_admin approval on creation

#### **2. Enrollment** (Athlete ↔ Academy)
- Many-to-many relationship with soft-delete safety
- Status: ACTIVE, SUSPENDED, LEFT
- When athlete leaves: soft-delete enrollment, stats remain via Sport-Athlete (separate service)

#### **3. Roster** (For Competition Submission)
- Collection of athletes from an academy for a specific competition
- Type: team | individual (based on sport config)
- Submitted to competition-service as "Participant"

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
