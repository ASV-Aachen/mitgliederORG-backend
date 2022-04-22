-- Data from ArbeitsstundenDB
CREATE OR REPLACE  VIEW arbeitsstundenmitglieder AS
SELECT first_name, last_name, email FROM user_
LEFT JOIN member member
ON member.user_id = user_.id;
