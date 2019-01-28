import time
import os
import sys


from swagger_client.rest import ApiException
import swagger_client.models
from pprint import pprint

#CLIENT=dict(endpoint="https://"+harbor_server+"/api")
ADMIN_CLIENT=dict(endpoint = "https://"+"192.168.17.100:30022"+"/apis")
USER_ROLE=dict(admin=0,normal=1)
TEARDOWN = True

def GetProductApi(username, password, harbor_server= "192.168.17.100:30022"):
    
    cfg = swagger_client.Configuration()
    cfg.host = "https://"+harbor_server+"/api"
    cfg.username = username
    cfg.password = password
    cfg.verify_ssl = False
    cfg.debug = True
    api_client = swagger_client.ApiClient(cfg)
    api_instance = swagger_client.DefaultApi(api_client)
    return api_instance
class TestResult(object):
    def __init__(self):
        self.num_errors = 0
        self.error_message = []
    def add_test_result(self, error_message):
        self.num_errors = self.num_errors + 1
        self.error_message.append(error_message)
    def get_final_result(self):
        if self.num_errors > 0:
            for each_err_msg in self.error_message:
                print ("Error message:", each_err_msg)
            raise Exception(r"Test case failed with {} errors.".format(self.num_errors))