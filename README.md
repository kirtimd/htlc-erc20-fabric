

## Hashed Time Lock Design
We'll store the tokens to be transferred as an agreement
for given lock_password -> get expiry_time
if expiry_time > current_time
    for lock_password_expiry_time -> get agreement(sender, receiver, amount) and transfer