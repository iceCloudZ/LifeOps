package ai.lifeops.family;

import java.util.List;

public record AiDraftParseResult(
        boolean ok,
        List<AiDraft> drafts,
        String failureReason
) {
    public static AiDraftParseResult ok(List<AiDraft> drafts) {
        return new AiDraftParseResult(true, drafts, null);
    }

    public static AiDraftParseResult failed(String reason) {
        return new AiDraftParseResult(false, List.of(), reason);
    }

    public boolean failed() {
        return !ok;
    }
}
