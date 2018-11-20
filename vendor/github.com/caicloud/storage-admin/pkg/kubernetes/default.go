package kubernetes

import (
	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitDefaultStorageTypes(cs Interface) error {
	tps := GetDefaultStorageTypes()
	for i := range tps {
		tp := &tps[i]
		_, e := cs.ResourceV1beta1().StorageTypes().Create(tp)
		if e != nil && !IsAlreadyExists(e) {
			return e
		}
	}
	return nil
}

func GetDefaultStorageTypes() []resv1b1.StorageType {
	return []resv1b1.StorageType{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "glusterfs",
			},
			Provisioner: StorageClassProvisionerGlusterfs,
			RequiredParameters: map[string]string{
				// "restauthenabled": "eg:\"true\"", // said deprecated, but in use now
				"restuserkey":     "eg:\"MxqrRIoe4ZP0ZwSeK+tK/48C\"", // deprecated, use here for make secrets
				"resturl":         "eg:\"http://1.2.3.4:5\"",
				"restuser":        "eg:\"admin\"",
				"secretNamespace": "eg:\"default\"",
				"secretName":      "eg:\"heketi-secret\"",
			},
			OptionalParameters: map[string]string{
				"clusterid":  "eg:\"630372ccdc720a92c681fb928f27b53f\"",
				"gidMin":     "eg:\"40000\"",
				"gidMax":     "eg:\"50000\"",
				"volumetype": "eg:\"replicate:3\"",
			},
		},
		//{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Name: "nas-netapp",
		//	},
		//	Provisioner: StorageClassProvisionerNetappNAS,
		//	RequiredParameters: map[string]string{
		//		"backendType": "eg:\"ontap-nas, ontap-nas-economy...\"",
		//	},
		//	OptionalParameters: map[string]string{
		//		"media":            "eg:\"hdd, hybrid, ssd\"",
		//		"provisioningType": "eg:\"thin, thick\"",
		//		"snapshots":        "eg:\"true, false\"",
		//		"clones":           "eg:\"true, false\"",
		//		"encryption":       "eg:\"true, false\"",
		//		"IOPS":             "eg:\"102400\"",
		//		"requiredStorage":  "eg:\"ontapnas_192.168.1.100:aggr1,aggr2;solidfire_192.168.1.101:bronze\"",
		//	},
		//},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nfs",
			},
			Provisioner: StorageClassProvisionerNFS,
			RequiredParameters: map[string]string{
				"server":     "eg:\"192.168.3.4\"",
				"exportPath": "eg:\"/export/nfs/data1\"",
			},
			OptionalParameters: map[string]string{
				"readOnly": "eg:\"true, false\"",
			},
		},
	}
}
