package ai.lifeops.family;

import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.notNullValue;

@QuarkusTest
class WebhookInboxResourceTest {

    @Test
    void acceptsWebhookMessageIntoFamilyInbox() {
        given()
                .header("X-LifeOps-Token", "dev-token")
                .contentType("application/json")
                .body("""
                        {
                          "source": "wechat",
                          "sender": "partner",
                          "content": "周五孩子要带彩笔，别忘了交水费。"
                        }
                        """)
                .when()
                .post("/api/inbox/webhook")
                .then()
                .statusCode(201)
                .body("id", notNullValue())
                .body("source", equalTo("wechat"))
                .body("sender", equalTo("partner"))
                .body("content", equalTo("周五孩子要带彩笔，别忘了交水费。"))
                .body("status", equalTo("new"));
    }
}
