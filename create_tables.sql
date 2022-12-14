DROP DATABASE IF EXISTS go;
CREATE DATABASE go;

DROP TABLE IF EXISTS Courses;
DROP TABLE IF EXISTS Coursedata;

\c go;

CREATE TABLE Courses (
    ID SERIAL,
    Coursecode VARCHAR(100) PRIMARY KEY
);

CREATE TABLE Coursedata (
    CourseID INT NOT NULL,
    CID VARCHAR(100),
    CNAME VARCHAR(100),
    CPREREQ VARCHAR(100)
);
