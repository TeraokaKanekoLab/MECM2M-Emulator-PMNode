import pymysql
from dotenv import load_dotenv
import os

load_dotenv()

host = "localhost"
user = os.getenv("MYSQL_USERNAME")
password = os.getenv("MYSQL_PASSWORD")
local_db = os.getenv("MYSQL_LOCAL_DB")
table = os.getenv("MYSQL_TABLE")

local_connection = pymysql.connect(host=host, user=user, password=password)
local_cursor = local_connection.cursor()

# データベースを新たに作成する
create_database_query = f"CREATE DATABASE IF NOT EXISTS {local_db};"
local_cursor.execute(create_database_query)

use_database_cursor_query = f"USE {local_db}"
local_cursor.execute(use_database_cursor_query)

# テーブル・インデックスを新たに作成する
create_table_query = f"CREATE TABLE IF NOT EXISTS {table}(PNodeID VARCHAR(20), Capability VARCHAR(20), Timestamp VARCHAR(30), Value DECIMAL(5,2), PSinkID VARCHAR(20), Lat DECIMAL(6,4), Lon DECIMAL(7,4));"
local_cursor.execute(create_table_query)
create_index_query = f"CREATE UNIQUE INDEX prim_index on {table}(PNodeID, Capability, Timestamp);"
local_cursor.execute(create_index_query)

# 変更をコミットして，接続を切断する
local_connection.commit()
local_cursor.close()
local_connection.close()

print("Create Local SensingDB")