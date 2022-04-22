-- CREATE OR REPLACE VIEW Mitglieder.`test_View_KeycloakAndWebsite` AS
-- SELECT sso.ID, sso.EMAIL, sso.EMAIL_VERIFIED, sso.ENABLED, sso.FIRST_NAME, sso.LAST_NAME, sso.USERNAME, website.id as websiteIDtest_View_KeycloakAndWebsite FROM KeycloackDB.USER_ENTITY sso
-- LEFT JOIN
-- websiteDB.auth_user website
-- ON
-- website.username = sso.USERNAME;

-- Data from Keycloak
CREATE OR REPLACE VIEW Mitglieder.`Keycloak` AS
SELECT sso.ID, sso.EMAIL, sso.EMAIL_VERIFIED, sso.ENABLED, sso.FIRST_NAME, sso.LAST_NAME, sso.USERNAME FROM KeycloackDB.USER_ENTITY sso;

-- Data from Webpage
CREATE OR REPLACE VIEW Mitglieder.`Website` AS
SELECT web.id, web.username, web.first_name, web.last_name, web.email, web.is_active, member.status 
FROM websiteDB.auth_user web
LEFT JOIN
websiteDB.member_profile member
ON
member.user_id = web.id;

-- Data from Bierkasse
-- TODO:
