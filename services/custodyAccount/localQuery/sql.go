package localQuery

var btcListQuery = `
select user.user_name ,sum(balance) as amount,'00' as asset_id
from(
    select user_account_balance_btc.amount as balance,user_account.user_id
    from user_account_balance_btc
    left join user_account on user_account_balance_btc.account_id = user_account.id
    where user_account_balance_btc.amount > 0
    union all

    select user_lock_balance.amount as balance,user_lock_account.user_id
    from user_lock_balance
    left join user_lock_account on user_lock_account.id = user_lock_balance.account_id
    where user_lock_balance.amount > 0 and user_lock_balance.asset_id = '00'
) as t
left join user on t.user_id = user.id
group by user_id
Order by amount DESC
limit ?
offset ?
`
var btcListQueryCT = `
select count(amount) as count,sum(amount) as total
from (
	select sum(balance) as amount
	from(
		select user_account_balance_btc.amount as balance,user_account.user_id
		from user_account_balance_btc
		left join user_account on user_account_balance_btc.account_id = user_account.id
		where user_account_balance_btc.amount > 0
		union all
	
		select user_lock_balance.amount as balance,user_lock_account.user_id
		from user_lock_balance
		left join user_lock_account on user_lock_account.id = user_lock_balance.account_id
		where user_lock_balance.amount > 0 and user_lock_balance.asset_id = '00'
	) as t
	group by user_id
) as num
`

var assetListQuery = `
select user.user_name ,sum(balance) as amount,? as asset_id
from(
    select user_account_balance.amount as balance,user_account.user_id
    from user_account_balance
    left join user_account on user_account_balance.account_id = user_account.id
    where user_account_balance.amount > 0 and user_account_balance.asset_id = ?
    union all

    select user_lock_balance.amount as balance,user_lock_account.user_id
    from user_lock_balance
    left join user_lock_account on user_lock_account.id = user_lock_balance.account_id
    where user_lock_balance.amount > 0 and user_lock_balance.asset_id = ?
) as t
left join user on t.user_id = user.id
group by user_id
Order by amount DESC
limit ?
offset ?
`
var assetListQueryCT = `
select count(amount) as count,sum(amount) as total,? as asset_id
from (
	select sum(balance) as amount
	from(
		select user_account_balance.amount as balance,user_account.user_id
		from user_account_balance
		left join user_account on user_account_balance.account_id = user_account.id
		where user_account_balance.amount > 0 and user_account_balance.asset_id = ?
		union all
	
		select user_lock_balance.amount as balance,user_lock_account.user_id
		from user_lock_balance
		left join user_lock_account on user_lock_account.id = user_lock_balance.account_id
		where user_lock_balance.amount > 0 and user_lock_balance.asset_id = ?
	) as t
left join user on t.user_id = user.id
group by user_id
) as num
`
