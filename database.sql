DROP TABLE IF EXISTS person;

CREATE TABLE person (
	id BIGSERIAL PRIMARY KEY,
	firstname text not null,
	lastname text not null,
	age integer not null
);

--insert into person (firstname, lastname)
--values ('terry', 'mcginnis');

select * from person;
