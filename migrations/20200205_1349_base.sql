create table files
(
	id varchar(128) not null
		primary key,
	internal_name text not null,
	uploaded_name text not null,
	created_at timestamp default CURRENT_TIMESTAMP not null,
	progress tinytext not null
);

