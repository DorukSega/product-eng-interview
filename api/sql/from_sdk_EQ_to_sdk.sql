-- from_sdk == to_sdk and not -1
select * 
from app 
where id in (select app_id from app_sdk where sdk_id = ? and installed = 1)
limit 10;