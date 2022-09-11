insert into public.transactions (playername, withdraw, deposit, currency, transactionref, completed)
values  ('player1', 400, 200, 'EUR', '1:UOwGgNHPgq3OkqRE', true),
        ('player1', 300, 0, 'EUR', '1:yYXJ6o8YFhhrhN29', true),
        ('player2', 0, 500, 'USD', '1:FxCcP9nC7Ua3aZxz', true);

insert into public.balances (playername, currency, balance)
values  ('player1', 'EUR', 500),
        ('player2', 'USD', 700),
        ('player3', 'EUR', 0);