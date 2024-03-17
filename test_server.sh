#!/bin/bash

sudo docker build -f test_script/Dockerfile -t test_script .
sudo docker run --network tp0_testing_net test_script:latest