# Winspire Core

Global loyalty & engagement platform — connecting brands, creators and consumers through games and rewards.

## 🚀 Quick Start

### Local Development (3 steps)

```bash
# 1. Start Supabase
cd platform/supabase && supabase start

# 2. Configure and start services
cd ../local
cp env.example .env  # Edit with Supabase credentials
make start           # or ./start.sh

# 3. Test
curl http://localhost/v1/cups
open http://localhost:8080  # Traefik dashboard
```

📖 Full guide: [platform/local/QUICK_START.md](platform/local/QUICK_START.md)

### Production Deployment

```bash
cd platform/terraform/environments/dev
terraform init
terraform apply
```

📖 Full guide: [platform/terraform/README.md](platform/terraform/README.md)

## Folder structure

[Structure](.cursor/rules/init_structure.mdc)

## Spawn new service

```python
pip install cookiecutter
cd services
cookiecutter template/
```

## License

This project is licensed under the **Business Source License 1.1 (BSL)**.  
Commercial use requires a separate license from Winspire Technologies.  
See the full license text here: [Business Source License 1.1](https://mariadb.com/bsl11/)