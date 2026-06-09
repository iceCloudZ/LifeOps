package com.lifeops.agent;

public interface DomainAgent {
    String domain();

    String systemPrompt();

    String retrieveContext(String query);
}
