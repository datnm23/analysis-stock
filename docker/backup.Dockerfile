FROM postgres:15-alpine

RUN apk add --no-cache bash aws-cli

COPY scripts/backup.sh /usr/local/bin/backup.sh
COPY scripts/restore.sh /usr/local/bin/restore.sh
RUN chmod +x /usr/local/bin/backup.sh /usr/local/bin/restore.sh

# Run backup daily at 2:00 AM ICT via crond
RUN echo "0 2 * * * /usr/local/bin/backup.sh >> /var/log/backup.log 2>&1" | crontab -

VOLUME ["/backups"]

CMD ["crond", "-f", "-l", "2"]
