-- Users
create table if not exists fio (
	id serial primary key,
    name varchar(30) not null,
    surname varchar(30) not null,
    patronymic varchar(30) default null,
    age integer not null,
    gender varchar(6) not null,
    nationality varchar(2) not null
);