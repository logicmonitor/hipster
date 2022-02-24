/*
 * Copyright 2018, Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.hipstershop;

import com.google.common.collect.ImmutableList;
import com.google.common.collect.ImmutableListMultimap;
import com.google.common.collect.Iterables;
import hipstershop.Demo.Ad;
import hipstershop.Demo.AdRequest;
import hipstershop.Demo.AdResponse;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.StatusRuntimeException;
import io.grpc.health.v1.HealthCheckResponse.ServingStatus;
import io.grpc.services.*;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Random;
import java.util.concurrent.TimeUnit;
import javax.sql.DataSource;
import org.apache.logging.log4j.Level;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.datasource.DriverManagerDataSource;

public final class AdService {

    private static final Logger logger = LogManager.getLogger(AdService.class);
    private static final AdService service = new AdService();
    private static final ImmutableListMultimap<String, Ad> adsMap = createAdsMap();
    private static final Random random = new Random();
    @SuppressWarnings("FieldCanBeLocal")
    private static int MAX_ADS_TO_SERVE = 2;
    private Server server;
    private HealthStatusManager healthMgr;
    private DataSource dataSource;

    public AdService() {

    }

    public AdService(DataSource dataSource) {
        this.dataSource = dataSource;
    }

    private static AdService getInstance() {
        return service;
    }

    private static ImmutableListMultimap<String, Ad> createAdsMap() {
        Ad camera = null, lens = null, recordPlayer = null, bike = null, baristaKit = null, airPlant = null, terrarium = null;
        DataSource dataSource = setUpDatabase();
        try {
            Connection con = dataSource.getConnection();
            String query = "select DISTINCT * from ads";
            Statement st = con.createStatement();
            ResultSet rs = st.executeQuery(query);
            while (rs.next()) {
                if (rs.getString("category").equals("photography")) {
                    if (rs.getString("product").equals("camera")) {
                        camera =
                            Ad.newBuilder()
                                .setRedirectUrl(rs.getString("redirectUrl"))
                                .setText(rs.getString("text"))
                                .build();
                    } else {
                        lens = Ad.newBuilder()
                            .setRedirectUrl(rs.getString("redirectUrl"))
                            .setText(rs.getString("text"))
                            .build();
                    }

                } else if (rs.getString("category").equals("vintage")) {
                    if (rs.getString("product").equals("recordPlayer")) {
                        recordPlayer =
                            Ad.newBuilder()
                                .setRedirectUrl(rs.getString("redirectUrl"))
                                .setText(rs.getString("text"))
                                .build();
                    }

                } else if (rs.getString("category").equals("cycling")) {
                    bike =
                        Ad.newBuilder()
                            .setRedirectUrl(rs.getString("redirectUrl"))
                            .setText(rs.getString("text"))
                            .build();

                } else if (rs.getString("category").equals("cookware")) {
                    baristaKit =
                        Ad.newBuilder()
                            .setRedirectUrl(rs.getString("redirectUrl"))
                            .setText(rs.getString("text"))
                            .build();
                } else {
                    if (rs.getString("product").equals("airPlant")) {
                        airPlant =
                            Ad.newBuilder()
                                .setRedirectUrl(rs.getString("redirectUrl"))
                                .setText(rs.getString("text"))
                                .build();
                    } else {
                        terrarium =
                            Ad.newBuilder()
                                .setRedirectUrl(rs.getString("redirectUrl"))
                                .setText(rs.getString("text"))
                                .build();
                    }
                }
            }
            logger.info("Getting data from the database");
        } catch (Exception e) {
            logger.info("Getting data from the code");
             camera =
                Ad.newBuilder()
                    .setRedirectUrl("/product/2ZYFJ3GM2N")
                    .setText("Film camera for sale. 50% off.")
                    .build();
             lens =
                Ad.newBuilder()
                    .setRedirectUrl("/product/66VCHSJNUP")
                    .setText("Vintage camera lens for sale. 20% off.")
                    .build();
             recordPlayer =
                Ad.newBuilder()
                    .setRedirectUrl("/product/0PUK6V6EV0")
                    .setText("Vintage record player for sale. 30% off.")
                    .build();
             bike =
                Ad.newBuilder()
                    .setRedirectUrl("/product/9SIQT8TOJO")
                    .setText("City Bike for sale. 10% off.")
                    .build();
             baristaKit =
                Ad.newBuilder()
                    .setRedirectUrl("/product/1YMWWN1N4O")
                    .setText("Home Barista kitchen kit for sale. Buy one, get second kit for free")
                    .build();
             airPlant =
                Ad.newBuilder()
                    .setRedirectUrl("/product/6E92ZMYYFZ")
                    .setText("Air plants for sale. Buy two, get third one for free")
                    .build();
              terrarium =
                Ad.newBuilder()
                    .setRedirectUrl("/product/L9ECAV7KIM")
                    .setText("Terrarium for sale. Buy one, get second one for free")
                    .build();
        }

        return ImmutableListMultimap.<String, Ad>builder()
            .putAll("photography", camera, lens)
            .putAll("vintage", camera, lens, recordPlayer)
            .put("cycling", bike)
            .put("cookware", baristaKit)
            .putAll("gardening", airPlant, terrarium)
            .build();
    }

    /**
     * Main launches the server from the command line.
     */
    public static void main(String[] args) throws IOException, InterruptedException, SQLException {
        logger.info("AdService starting.");
        System.out.println("M running successfuly");
        final AdService service = getInstance();
        service.start();
        service.blockUntilShutdown();
    }

    private static DataSource setUpDatabase() {

        String driverClassName = "com.mysql.jdbc.Driver";
        String username = "root";
        String password = "";
        String url = "jdbc:mysql://localhost:3306/hipstershop";

        DriverManagerDataSource dataSource = new DriverManagerDataSource();
        dataSource.setDriverClassName(driverClassName);
        System.out.println(driverClassName);
        System.out.println(url);
        dataSource.setUrl(url);
        dataSource.setUsername(username);
        dataSource.setPassword(password);

        return dataSource;
    }

    private Collection<Ad> getAdsByCategory(String category) {
        return adsMap.get(category);
    }

    private void start() throws IOException {
        int port = Integer.parseInt(System.getenv().getOrDefault("PORT", "9555"));
        healthMgr = new HealthStatusManager();

        server =
            ServerBuilder.forPort(port)
                .addService(new AdServiceImpl())
                .addService(healthMgr.getHealthService())
                .build()
                .start();
        logger.info("Ad Service started, listening on " + port);
        Runtime.getRuntime()
            .addShutdownHook(
                new Thread(
                    () -> {
                        // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                        System.err.println(
                            "*** shutting down gRPC ads server since JVM is shutting down");
                        AdService.this.stop();
                        System.err.println("*** server shut down");
                    }));
        healthMgr.setStatus("", ServingStatus.SERVING);
    }

    private void stop() {
        if (server != null) {
            healthMgr.clearStatus("");
            server.shutdown();
        }
    }

    private List<Ad> getRandomAds() {
        List<Ad> ads = new ArrayList<>(MAX_ADS_TO_SERVE);
        Collection<Ad> allAds = adsMap.values();
        for (int i = 0; i < MAX_ADS_TO_SERVE; i++) {
            ads.add(Iterables.get(allAds, random.nextInt(allAds.size())));
        }
        return ads;
    }

    /**
     * Await termination on the main thread since the grpc library uses daemon threads.
     */
    private void blockUntilShutdown() throws InterruptedException {
        if (server != null) {
            server.awaitTermination();
        }
    }

    private static class AdServiceImpl extends hipstershop.AdServiceGrpc.AdServiceImplBase {

        /**
         * Retrieves ads based on context provided in the request {@code AdRequest}.
         *
         * @param req the request containing context.
         * @param responseObserver the stream observer which gets notified with the value of {@code
         * AdResponse}
         */
        @Override
        public void getAds(AdRequest req, StreamObserver<AdResponse> responseObserver) {
            try {
                List<Ad> allAds = chooseAds(req);
                AdResponse reply = AdResponse.newBuilder().addAllAds(allAds).build();
                responseObserver.onNext(reply);
                responseObserver.onCompleted();
            } catch (StatusRuntimeException e) {
                logger.log(Level.WARN, "GetAds Failed with status {}", e.getStatus());
                responseObserver.onError(e);
            }
        }

        private List<Ad> chooseAds(AdRequest req) {
            AdService service = AdService.getInstance();
            List<Ad> allAds = new ArrayList<>();
            logger.info("received ad request (context_words=" + req.getContextKeysList() + ")");

            if (req.getContextKeysCount() > 0) {
                for (int i = 0; i < req.getContextKeysCount(); i++) {
                    Collection<Ad> ads = service.getAdsByCategory(req.getContextKeys(i));
                    allAds.addAll(ads);
                }
            } else {
                allAds = service.getRandomAds();
            }
            if (allAds.isEmpty()) {
                allAds = service.getRandomAds();
            }
            return allAds;
        }
    }

}