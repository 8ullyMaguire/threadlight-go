# Maintenance

This page covers common maintenance tasks for keeping your Threadlight server healthy.

## Database Migrations

Threadlight automatically runs database migrations on startup, so you don't need to do anything manually. When you update Threadlight to a newer version, the new migrations will run automatically the next time you restart the server.

## Content Pruning

Old and low-quality content is automatically cleaned up by the **prune worker**. This runs once an hour and removes:

- Posts older than 6 months that have fewer than 3 likes
- The age threshold and minimum likes are configurable in the code

This keeps the database from growing too large with stale content.

## Backing Up the Database

Regular backups are essential. Here's a simple backup script:

```bash
#!/bin/bash
BACKUP_DIR=/opt/polaris/backups
mkdir -p $BACKUP_DIR
pg_dump polaris > $BACKUP_DIR/polaris-$(date +%Y%m%d-%H%M%S).sql
# Keep only the last 7 days of backups
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
```

Set it to run daily with cron:

```bash
crontab -e
# Add this line to run at 3 AM every day:
0 3 * * * /opt/polaris/backup.sh
```

## Restoring from a Backup

```bash
# Create a fresh database
sudo -u postgres psql -c "DROP DATABASE polaris;"
sudo -u postgres psql -c "CREATE DATABASE polaris OWNER polaris;"

# Restore the backup
psql polaris < /opt/polaris/backups/polaris-20250101-030000.sql
```

## Monitoring

Keep an eye on:

- **Disk space**: Database and image uploads can grow quickly
- **Memory usage**: Redis and PostgreSQL use memory for caching
- **CPU usage**: Heavy activity or media processing can spike CPU

You can use tools like `htop`, `df -h`, and `journalctl -u polaris` to check the health of your server.

## Logs

Threadlight logs to standard output (stdout). If you're using systemd, you can view logs with:

```bash
journalctl -u polaris -f
```

## Updating Threadlight

To update to the latest version:

```bash
cd /opt/polaris
git pull
cd web && pnpm install && pnpm build && cd ..
go build -o polaris ./cmd/polaris
sudo systemctl restart polaris
```

The database migrations will run automatically on restart. Make sure to back up the database before updating.
