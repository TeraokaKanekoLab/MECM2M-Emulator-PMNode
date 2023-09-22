import pymysql
from dotenv import load_dotenv
import os

load_dotenv()

host = "localhost"
user = os.getenv("MYSQL_USERNAME")
password = os.getenv("MYSQL_PASSWORD")
local_db = os.getenv("MYSQL_LOCAL_DB")
table = os.getenv("MYSQL_TABLE")

# DBに接続
local_connection = pymysql.connect(host=host, user=user, password=password, database=local_db)
local_cursor = local_connection.cursor()

# テーブル内の全データを削除するクエリを実行
delete_query = f"DELETE FROM {table};"
local_cursor.execute(delete_query)

# テーブルを削除する
drop_table_query = f"DROP TABLE {table}"
local_cursor.execute(drop_table_query)

# 変更をコミットして，接続を切断する
local_connection.commit()
local_cursor.close()
local_connection.close()

print("Clear Local SensingDB")