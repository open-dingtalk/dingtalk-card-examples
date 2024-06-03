package com.card.java;

import com.dingtalk.open.app.api.OpenDingTalkClient;
import com.dingtalk.open.app.api.OpenDingTalkStreamClientBuilder;
import com.dingtalk.open.app.api.callback.DingTalkStreamTopics;
import com.dingtalk.open.app.api.security.AuthClientCredential;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;

@Component
public class Listener {
    @Autowired
    private ChatBotHandler chatBotHandler;
    @Autowired
    private CardCallbackHandler cardCallbackHandler;

    @Value("${dingtalk.app.client-id}")
    private String clientId;
    @Value("${dingtalk.app.client-secret}")
    private String clientSecret;

    @PostConstruct
    public void init() throws Exception {
        // init stream client
        OpenDingTalkClient client = OpenDingTalkStreamClientBuilder
                .custom()
                .credential(new AuthClientCredential(clientId, clientSecret))
                .registerCallbackListener(DingTalkStreamTopics.BOT_MESSAGE_TOPIC, chatBotHandler)
                .registerCallbackListener(DingTalkStreamTopics.CARD_CALLBACK_TOPIC, cardCallbackHandler)
                .build();
        client.start();
    }
}
