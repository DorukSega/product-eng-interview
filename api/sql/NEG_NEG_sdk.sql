-- Negative value to Negative Value
select * 
from app 
where id in (
               select app_id from app_sdk where sdk_id not in (?)
            )
limit 10;