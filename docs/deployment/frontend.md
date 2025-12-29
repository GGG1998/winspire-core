# Digitalocean

**Admin**:
- gabriel.domanowski@gowinspire.com

**Envirnonemnt**:
- dev

## Dev

**Url**: https://winspire-dev-s63lr.ondigitalocean.app/auth/user/login

### Deployment Proces

1. Create PR

Create a Pull Request targeting the winspire-app/dev branch:
https://github.com/Winspire-Lab/winspire-core/compare/winspire-app/dev...<your-branch>

2. Code Review

- PR must be reviewed and approved by a team member
- All CI checks must pass

3. Merge & Auto-Deploy

- Once approved, merge the PR
- Deployment to dev environment triggers automatically after merge

That's it - the pipeline handles the rest.