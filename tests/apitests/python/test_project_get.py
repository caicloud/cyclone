from __future__ import absolute_import

import unittest


from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project


class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project= project


    @classmethod
    def tearDown(self):
        print ("Case completed")

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by admin;
        kk = {"x_tenant":"devops"}
        data = self.project.get_projects(kk)
        print(data)

     
if __name__ == '__main__':
    unittest.main()