package com.lifeops;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
@MapperScan("com.lifeops.mapper")
public class LifeOpsApplication {
    public static void main(String[] args) {
        SpringApplication.run(LifeOpsApplication.class, args);
    }
}
