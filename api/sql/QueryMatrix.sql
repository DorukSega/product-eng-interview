select
from_sdk,
case 
	when sdk_id not in (?) then -1
	else sdk_id
end as to_sdk, 
COUNT(*) as count
from app_sdk as M
inner join 
	(select 
		 case 
			 when sdk_id not in (?) then -1 
			 else sdk_id 
		 end as from_sdk,
		 app_id, installed from app_sdk) as C
on C.app_id = M.app_id 
where M.installed = 1
and ((C.installed = 0 and from_sdk != to_sdk) or (C.installed =1 and from_sdk = to_sdk and from_sdk != -1))
group by to_sdk, from_sdk
union 
select -1 as from_sdk, -1 as to_sdk, COUNT(*) as count from app where id in (select distinct app_id from app_sdk where sdk_id not in (?));