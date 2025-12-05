// Define an environment named "local"
env "local" {
  // Declare where the schema definition resides.
  src = "file://migrations"

  // Define the URL of the database which is managed
  // in this environment.
  url = getenv("DATABASE_URL")

  // Define the URL of the Dev Database for this environment
  dev = "docker://postgres/15/dev?search_path=public"

  migration {
    // URL where the migration directory resides.
    dir = "file://migrations"
  }
}

