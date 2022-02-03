create table "names" (
	"id" integer primary key,
	"name" text not null unique
);

create table "combinations" (
	"id" integer primary key,
	"ip" integer not null,
	"name" integer not null references "names"("id"),
	unique ("ip", "name")
);

create index idx_comb_name on "combinations"("name");
create index idx_comb_ip on "combinations"("ip");

create table "servers" (
	"id" integer primary key,
	"ip" text not null,
	"port" integer not null,
	"description" text not null,
	"protocol" integer not null,
	"mod" integer,
	"last_seen" integer not null default (strftime('%s', 'now')),
	unique ("ip", "port", "protocol")
);

create index idx_srv_descr on "servers"("description");

create table "sightings" (
	"combination" integer not null references "combinations"("id"),
	"server" integer not null references "servers"("id"),
	"timestamp" integer not null default (strftime('%s', 'now')),
	unique ("combination", "timestamp")
);

create table "games" (
	"id" integer primary key,
	"server" integer not null references "servers"("id"),
	"master_mode" integer not null,
	"game_mode" integer not null,
	"map" text not null,
	"started_at" integer not null,
	"seconds_left" integer not null,
	"last_recorded_at" integer not null default (strftime('%s', 'now')),
	"ended_at" integer,
	unique ("server", "started_at")
);

create table "scores" (
	"game" integer not null references "games"("id"),
	"team" text not null,
	"points" integer not null,
	unique ("game", "team")
);

create table "stats" (
	"combination" integer not null references "combinations"("id"),
	"game" integer not null references "games"("id"),
	"team" text,
	"frags" integer,
	"deaths" integer,
	"accuracy" integer,
	"teamkills" integer,
	"flags" integer,
	"recorded_at" integer not null default (strftime('%s', 'now'))
);
