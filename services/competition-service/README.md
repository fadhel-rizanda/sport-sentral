## **Competition SERVICE IN DETAIL**
### **Key Responsibilities**
```
┌────────────────────────────────────────────────────────────┐
│         COMPETITION SERVICE                                │
│                                                            │
│  ├─ Competition (Tournament/League)                        │
│  │  ├─ SportID (FK → sport-service)                        │
│  │  ├─ HostAcademyID (FK → academy-service)                │
│  │  └─ Tier (OFFICIAL, REGIONAL, LOCAL, UNRATED)           │
│  │                                                         │
│  ├─ Competition Branch (Hierarchy)                         │
│  │  ├─ ParentBranchID (for sub-divisions)                  │
│  │  └─ Name (e.g., "Division A", "U-17 Category")          │
│  │                                                         │
│  ├─ Match (Individual games/events)                        │
│  │  ├─ BranchID (FK → CompetitionBranch)                   │
│  │  ├─ HomeTeamID, AwayTeamID (FK → Roster)                │
│  │  ├─ ScheduledAt, StartedAt, EndedAt                     │
│  │  └─ Status (SCHEDULED, LIVE, COMPLETED, CANCELLED)      │
│  │                                                         │
│  ├─ Match Stats (Pivot table - all stats per athlete)      │
│  │  ├─ MatchID (FK)                                        │
│  │  ├─ AthleteID (FK → athlete via enrollment)             │
│  │  ├─ StatTypeID (FK → replicated_tags with STAT_*)       │
│  │  └─ Value (numeric value)                               │
│  │                                                         │
│  └─ Athlete Stats Aggregate (Pivot table - avg stats)      │
│     ├─ CompetitionID (FK)                                  │
│     ├─ AthleteID (FK)                                      │
│     ├─ StatTypeID (FK)                                     │
│     ├─ TotalMatches                                        │
│     ├─ AvgValue, MaxValue, MinValue                        │
│     └─ UpdatedAt (denormalized for quick fetch)            │
│                                                            │
└────────────────────────────────────────────────────────────┘
```