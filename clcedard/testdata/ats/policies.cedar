// Users that are owners of the special management organization
// "crewlinker" are allows to perform super-admin actions.
permit (
    principal,
    action in Ats::Action::"gSuperAdminActions",
    resource
)
when
{
    principal.ownerOfOrganizations.contains(Ats::Organization::"crewlinker") &&
    context.source_ip.isInRange(ip("222.222.222.0/24"))
};