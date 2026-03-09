#!/bin/bash
# MySQL 主从复制初始化脚本
# 在 docker-compose up 后执行
# Usage: bash deploy/mysql/init-replication.sh

set -e

echo "=== 等待 MySQL Master 就绪 ==="
until docker exec journal-mysql-master mysqladmin ping -h localhost -u root -proot123456 --silent 2>/dev/null; do
    echo "等待 master..."
    sleep 2
done

echo "=== 等待 MySQL Replica1 就绪 ==="
until docker exec journal-mysql-replica1 mysqladmin ping -h localhost -u root -proot123456 --silent 2>/dev/null; do
    echo "等待 replica1..."
    sleep 2
done

echo "=== 等待 MySQL Replica2 就绪 ==="
until docker exec journal-mysql-replica2 mysqladmin ping -h localhost -u root -proot123456 --silent 2>/dev/null; do
    echo "等待 replica2..."
    sleep 2
done

echo "=== 在 Master 上创建复制用户 ==="
docker exec journal-mysql-master mysql -uroot -proot123456 -e "
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED WITH mysql_native_password BY 'repl_password';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;
"

echo "=== 获取 Master GTID 状态 ==="
MASTER_STATUS=$(docker exec journal-mysql-master mysql -uroot -proot123456 -e "SHOW MASTER STATUS\G" 2>/dev/null)
echo "$MASTER_STATUS"

echo "=== 配置 Replica1 ==="
docker exec journal-mysql-replica1 mysql -uroot -proot123456 -e "
STOP SLAVE;
RESET SLAVE ALL;
CHANGE MASTER TO
    MASTER_HOST='mysql-master',
    MASTER_USER='repl',
    MASTER_PASSWORD='repl_password',
    MASTER_AUTO_POSITION=1;
START SLAVE;
"

echo "=== 配置 Replica2 ==="
docker exec journal-mysql-replica2 mysql -uroot -proot123456 -e "
STOP SLAVE;
RESET SLAVE ALL;
CHANGE MASTER TO
    MASTER_HOST='mysql-master',
    MASTER_USER='repl',
    MASTER_PASSWORD='repl_password',
    MASTER_AUTO_POSITION=1;
START SLAVE;
"

echo "=== 检查 Replica1 状态 ==="
docker exec journal-mysql-replica1 mysql -uroot -proot123456 -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep -E "(Slave_IO_Running|Slave_SQL_Running|Seconds_Behind_Master)"

echo "=== 检查 Replica2 状态 ==="
docker exec journal-mysql-replica2 mysql -uroot -proot123456 -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep -E "(Slave_IO_Running|Slave_SQL_Running|Seconds_Behind_Master)"

echo ""
echo "✅ MySQL 主从复制配置完成！"
echo "   Master:   127.0.0.1:13306"
echo "   Replica1: 127.0.0.1:13307"
echo "   Replica2: 127.0.0.1:13308"
