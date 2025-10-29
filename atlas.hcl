env "local" {
  url = "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable"
  dev = "docker://postgres/16/test?search_path=public"
  migration {
    dir = "file://ent/migrate/migrations"
  }
}
