From imega/mysql-client:latest
COPY backup.sh ./
COPY store.sh ./
CMD ["/bin/bash", "-c", "backup.sh"]
