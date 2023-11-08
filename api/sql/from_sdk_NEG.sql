-- from_sdk to negative value
select * 
from app 
where id in (
                select T.app_id from app_sdk as T 
                inner join
                (select app_id from app_sdk where sdk_id = ? and installed = 0) as F 
                on F.app_id = T.app_id
                where T.sdk_id not in (?) and T.installed = 1
            )
limit 10;