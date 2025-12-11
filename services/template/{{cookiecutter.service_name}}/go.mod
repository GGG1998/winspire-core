module {{ cookiecutter.go_module }}

go 1.25

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joho/godotenv v1.5.1
{% if cookiecutter.use_redis == "true" %}
	github.com/redis/go-redis/v9 v9.7.3
{% endif %}
	github.com/winspire/libs/go/auth v0.0.0
)

replace github.com/winspire/libs/go/auth => ../../libs/go/auth


