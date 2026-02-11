# Disclaimer

This repository is a mirror of the original Winspire Core repository. The original repository is located at https://github.com/winspire-core/winspire-core.

MVP was created in rapid environment so most code wasn't updated and clear and is full of security leaks.

I left code as a proof of evidence of rapid hypothesis testing in a startup environment. Most important is the fact we tested our hipotesis and decided it's not worth to continue.

- Hard cooperation with streamers
- Development of games is time and money consuming
- People is easier engaged to visit venue without gaming

According to effort maintaining microservices and work with transactions between events
I dicided to move to monolith architecture and make a mad, so in the code
you can see events and soap way.

# What I've done so far

Here is video demo: [https://youtu.be/Fk3bM2tSfek](https://youtu.be/Fk3bM2tSfek)

- Auth using supabase
- Two types user: Host and Player
- CRUD for tournament(
    states: draft, 
    registration_open, 
    registration_closed, 
    started, 
    completed,
    cancelled
)
- (The part still exist) Event-Driven architecture for managing tournament state
    - Generating brackets
    - Swap players between match
    - lobby, pre-lobby
    - Selection of the winner
- Loading games from S3
- Mini-panel for managing games(uploading to S3)
- Not sure wheter CI/CD fully work but something should
- Terraform for deploying infrastructure

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
