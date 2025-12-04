// Atlas configuration for matchmaking service migrations

// Environment variables
variable "database_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

// Define the environment
env "local" {
  src = "file://migrations"
  url = var.database_url

  // Migration directory configuration
  migration {
    dir = "file://migrations"
  }

  // Format for migration files
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "dev" {
  src = "file://migrations"
  url = var.database_url

  migration {
    dir = "file://migrations"
  }
}

env "prod" {
  src = "file://migrations"
  url = var.database_url

  migration {
    dir = "file://migrations"
    // Baseline version for production safety
    baseline = "000001_create_brackets_table"
  }

  // Require explicit confirmation for production
  diff {
    skip {
      drop_schema = true
      drop_table  = true
    }
  }
}

