import pymysql
import pymysql.cursors

connection = pymysql.connect(host='127.0.0.1',
                             user='root',
                             password='',
                             db='mydb',
                             cursorclass=pymysql.cursors.DictCursor)

try:
    with connection.cursor() as cursor:
        sql = "SHOW tables"
        cursor.execute(sql)
        rows = cursor.fetchall()
        for row in rows:
            print(row)
finally:
    connection.close()
