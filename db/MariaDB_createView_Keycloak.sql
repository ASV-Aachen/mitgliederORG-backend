-- Data from Keycloak
CREATE OR REPLACE VIEW Mitglieder.`Keycloak` AS SELECT sso.ID, sso.EMAIL, sso.EMAIL_VERIFIED, sso.ENABLED, sso.FIRST_NAME, sso.LAST_NAME, sso.USERNAME FROM KeycloackDB.USER_ENTITY sso;
