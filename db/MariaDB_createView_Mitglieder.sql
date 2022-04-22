-- Data from Webpage
CREATE OR REPLACE VIEW Mitglieder.`Website` AS SELECT web.id, web.username, web.first_name, web.last_name, web.email, web.is_active, member.status FROM websiteDB.auth_user web LEFT JOIN websiteDB.member_profile member ON member.user_id = web.id;
