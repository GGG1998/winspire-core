// Atlas configuration for matchmaking service
env "local" {
  src = "file://migrations"
  url = "postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable"

  migration {
    dir = "file://migrations"
    revisions_schema = "matchmaking_revisions"
  }
}

env "dev" {
  src = "file://migrations"

  migration {
    dir = "file://migrations"
    revisions_schema = "matchmaking_revisions"
  }
}

env "staging" {
  src = "file://migrations"

  migration {
    dir = "file://migrations"
    revisions_schema = "matchmaking_revisions"
  }
}

env "prod" {
  src = "file://migrations"

  migration {
    dir = "file://migrations"
    revisions_schema = "matchmaking_revisions"
  }
}
