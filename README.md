# cockroachdb-omnibox



```sql
CREATE DATABASE omnibox; 
```

```sql
CREATE USER omnibox0;
CREATE USER omnibox1;
CREATE USER omnibox1;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE datapoints TO omnibox0, omnibox1, omnibox2;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE message_board TO omnibox0, omnibox1, omnibox2;
```

```sql
SHOW CLUSTER SETTING server.host_based_authentication.configuration;
```


```sql
SET CLUSTER SETTING server.host_based_authentication.configuration = '
host all omnibox0 127.0.0.1/32 trust
host all omnibox0 ::1/128 trust
host all omnibox1 127.0.0.1/32 trust
host all omnibox1 ::1/128 trust
host all omnibox2 127.0.0.1/32 trust
host all omnibox2 ::1/128 trust
host all all all password';
```


```sql
SELECT * FROM (SELECT device_from, device_to, received_at, LEFT(message,30), read_at FROM message_board ORDER BY received_at DESC  LIMIT 20) sub ORDER BY received_at; 
```

```sql
CREATE CHANGEFEED WITH format=json AS 
SELECT device_from, device_to, received_at, message, read_at
FROM message_board;   
```

