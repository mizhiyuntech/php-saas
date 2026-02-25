# AGENTS.md

## Cursor Cloud specific instructions

### Overview

This is a PHP SaaS application repository (`php-saas`). As of initial setup, the repository contains only a `README.md` — no application code, framework, or `composer.json` has been added yet.

### Environment

- **PHP**: 8.3.6 (installed via `apt`: `php`, `php-cli`, and common extensions including mbstring, xml, curl, zip, mysql, sqlite3, bcmath, tokenizer, intl)
- **Composer**: 2.9.5 (installed globally at `/usr/local/bin/composer`)
- **Built-in dev server**: `php -S localhost:8080` (use from the project root once application code exists)

### Development commands

Once application code and a `composer.json` are added:

- **Install dependencies**: `composer install --no-interaction`
- **Run dev server**: `php -S localhost:8080 -t public/` (adjust `-t` to the document root)
- **Run tests**: `composer test` or `./vendor/bin/phpunit` (once PHPUnit is configured)
- **Lint**: `composer lint` or `./vendor/bin/phpcs` (once PHP_CodeSniffer is configured)

### Caveats

- The VM does not have a database server (MySQL/PostgreSQL) pre-installed. If the application requires one, install it or use SQLite for development.
- Apache is installed as a side effect of the `php` package but is not running; use PHP's built-in server for development instead.
