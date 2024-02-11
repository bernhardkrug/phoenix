create table if not exists ddhub.locations
(
    country_isocode varchar(255) not null,
    region_isocode  varchar(255) not null,
    primary key (country_isocode, region_isocode)
);

create table if not exists ddhub.employees
(
    email           varchar(255) not null,
    first_name      varchar(255),
    full_name       varchar(255),
    last_name       varchar(255),
    acronym         varchar(255),
    active          boolean default true,
    locked          boolean default false,
    visible         boolean default true,
    country_isocode varchar(255),
    region_isocode  varchar(255),
    birthday        varchar(10),
    primary key (email),
    constraint employee_location
        foreign key (country_isocode, region_isocode) references ddhub.locations (country_isocode, region_isocode)
);

create table if not exists ddhub.usergroups
(
    name varchar(255) not null,
    primary key (name)
);

create table if not exists ddhub.employee_usergroup
(
    name  varchar(255) not null,
    email varchar(255) not null,
    primary key (name, email),
    constraint employee_usergroup_usergroup
        foreign key (name) references ddhub.usergroups (name),
    constraint employee_usergroup_employee
        foreign key (email) references ddhub.employees (email)
);

create table if not exists ddhub.labels
(
    id           varchar(255) not null,
    display_name varchar(255) not null unique,
    primary key (id)
);

CREATE INDEX idx_displayName
    ON ddhub.labels (display_name);

create table if not exists ddhub.employee_labels
(
    label_id varchar(255) not null,
    email    varchar(255) not null,
    primary key (label_id, email),
    constraint employee_labels_user
        foreign key (email) references ddhub.employees (email),
    constraint employee_labels_labels
        foreign key (label_id) references ddhub.labels (id)
);

create table if not exists ddhub.permissions
(
    resource     varchar(255) not NULL,
    access_level varchar(255) not NULL,
    primary key (resource, access_level)
);

create table if not exists ddhub.usergroup_permission
(
    resource     varchar(255) not null,
    access_level varchar(255) not null,
    user_group   varchar(255) not null,
    primary key (user_group, resource, access_level),
    constraint userGroups_access_accessLevel
        foreign key (resource, access_level) references ddhub.permissions (resource, access_level),
    constraint userGroups_access_userGroup
        foreign key (user_group) references ddhub.usergroups (name)
);