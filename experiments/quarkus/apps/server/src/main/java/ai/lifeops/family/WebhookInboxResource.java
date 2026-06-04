package ai.lifeops.family;

import jakarta.inject.Inject;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.HeaderParam;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;

@Path("/api/inbox/webhook")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class WebhookInboxResource {

    private static final String DEV_TOKEN = "dev-token";

    @Inject
    InboxService inboxService;

    @POST
    public Response create(@HeaderParam("X-LifeOps-Token") String token, WebhookInboxRequest request) {
        if (!DEV_TOKEN.equals(token)) {
            return Response.status(Response.Status.UNAUTHORIZED).build();
        }
        return Response.status(Response.Status.CREATED)
                .entity(inboxService.create(request))
                .build();
    }
}
