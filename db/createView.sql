CREATE VIEW Mitglieder.`test_View_KeycloakAndWebsite` AS
SELECT sso.ID, sso.EMAIL, sso.EMAIL_VERIFIED, sso.ENABLED, sso.FIRST_NAME, sso.LAST_NAME, sso.USERNAME, website.id as websiteIDtest_View_KeycloakAndWebsite FROM KeycloackDB.USER_ENTITY sso
LEFT JOIN
websiteDB.auth_user website
ON
website.username = sso.USERNAME;
