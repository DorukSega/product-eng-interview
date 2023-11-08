--Stripe has 10,572 apps currently using their SDK (of the sampled apps)
--Stripe churned 22 customers to PayPal
--Stripe churned 8,082 customers to another solution not covered in this matrix
--PayPal acquired 11,844 app integrations from another solution not covered in this matrixs
--matrix (33,2081,875)
select
    from_sdk,
    case 
        when sdk_id not in (33,2081,875) then -1
        else sdk_id
    end as to_sdk, 
    COUNT(*) as count
from app_sdk as M
    inner join 
        (select 
             case 
                 when sdk_id not in (33,2081,875) then -1 
                 else sdk_id 
             end as from_sdk,
             app_id, installed from app_sdk) as C
    on C.app_id = M.app_id 
where M.installed = 1
and ((C.installed = 0 and from_sdk != to_sdk) or (C.installed =1 and from_sdk = to_sdk and from_sdk != -1))
group by to_sdk, from_sdk;
--879,587 apps from the sample haven't integrated any of these three payments SDKs
select COUNT(*) from app where id not in (select distinct app_id from app_sdk where sdk_id in (33,2081,875));

select
    from_sdk,
    case 
        when sdk_id not in (33) then -1
        else sdk_id
    end as to_sdk, 
    COUNT(*) as count
from app_sdk as M
    inner join 
        (select 
             case 
                 when sdk_id not in (33) then -1 
                 else sdk_id 
             end as from_sdk,
             app_id, installed from app_sdk) as C
    on C.app_id = M.app_id 
where M.installed = 1
and ((C.installed = 0 and from_sdk != to_sdk) or (C.installed =1 and from_sdk = to_sdk and from_sdk != -1))
group by to_sdk, from_sdk
union 
select -1 as from_sdk, -1 as to_sdk, COUNT(*) from app where id not in (select distinct app_id from app_sdk where sdk_id in (33));