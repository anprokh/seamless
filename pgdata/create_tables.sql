create table if not exists transactions
(
  id             bigserial primary key,
  playername     varchar(255) NOT NULL,
  withdraw       int default 0,
  deposit        int default 0,
  currency       varchar(3),
  transactionref varchar(255) NOT NULL,
  completed      boolean default FALSE,
  canceled       boolean default FALSE
);
create index ix_transactionref on transactions (transactionref);


create table if not exists balances
(
  id             bigserial primary key,
  playername     varchar(255) UNIQUE NOT NULL,
  currency       varchar(3),
  balance        int default 0
);
create index ix_playername on balances (playername);