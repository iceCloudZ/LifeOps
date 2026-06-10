package com.lifeops.agent;

import dev.langchain4j.service.TokenStream;
import dev.langchain4j.service.UserMessage;

public interface ButlerAssistant {
    TokenStream chat(@UserMessage String message);
}
