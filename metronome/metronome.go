package metronome

var metronomeTracks = map[string]map[string]string{
	"60": {
		"4/4": "CQACAgIAAxkBAAEB0tZiLltEwLNCmr5IdNYRwCIWL9oqDQACbRQAAo5qeElNbVbX69LlbyME",
	},
	"61": {
		"4/4": "CQACAgIAAxkBAAEB0thiLltHeK5UBF7ElMQbWZLWPdCM_gACbhQAAo5qeEl4CKM_dx9_RSME",
	},
	"62": {
		"4/4": "CQACAgIAAxkBAAEB0tpiLltK8ZcLUdTecFP_esByyZWogAACbxQAAo5qeElsVVOZfgweHCME",
	},
	"63": {
		"4/4": "CQACAgIAAxkBAAEB0txiLltNG48tM9jOIZUlo32mXjvwDgACcBQAAo5qeEkneWFpuzUF8iME",
	},
	"64": {
		"4/4": "CQACAgIAAxkBAAEB0t5iLltPa_0WwbVguU0QIf-K7XQ_5gACcRQAAo5qeEkXaKSpMQYrPSME",
	},
	"65": {
		"4/4": "CQACAgIAAxkBAAEB0uBiLltSI-Ls-AShVthdDP_vDwgsegACchQAAo5qeElxzPM5hqCxlCME",
	},
	"66": {
		"4/4": "CQACAgIAAxkBAAEB0uJiLltWSeRYxSZq2zG7rXhw9f816wACcxQAAo5qeEkpTReg49PO1SME",
	},
	"67": {
		"4/4": "CQACAgIAAxkBAAEB0uRiLltZfSdMHiPPOuu79_yNp-Q0MAACdBQAAo5qeEmOdeDFi8AGoSME",
	},
	"68": {
		"4/4": "CQACAgIAAxkBAAEB0uZiLltdOvyhbqTohCYCsLPtQsjhaQACdRQAAo5qeElpV5AsqXks-CME",
	},
	"69": {
		"4/4": "CQACAgIAAxkBAAEB0uhiLlt2ue4iAAFfN__5RkMMw2we_qQAAnYUAAKOanhJsQ171ZasGvgjBA",
	},
	"70": {
		"4/4": "CQACAgIAAxkBAAEB0upiLluO8itPVjbxEgKQsReNm148LwACdxQAAo5qeElwVg3c1j_LHCME",
	},
	"71": {
		"4/4": "CQACAgIAAxkBAAEB0uxiLlumkoDY7JRcn1-PpekHSXgWEgACeBQAAo5qeEnAT7qwnliZ6iME",
	},
	"72": {
		"4/4": "CQACAgIAAxkBAAEB0u5iLlu-n6F0luAdSm9_B7zWwNluWAACeRQAAo5qeElUJmwZbJHA0yME",
	},
	"73": {
		"4/4": "CQACAgIAAxkBAAEB0vBiLlvVsYVOwMg4EpsslDOuHV9s1AACehQAAo5qeEk6HUCFB_qZbyME",
	},
	"74": {
		"4/4": "CQACAgIAAxkBAAEB0vJiLlvtk1UZzeRLDlEO3Y54mHlLdAACexQAAo5qeElwfBerhXuUGiME",
	},
	"75": {
		"4/4": "CQACAgIAAxkBAAEB0vRiLlwCzl5-MAujq6BJNYpTIkf5LwACfBQAAo5qeEnw0XimiFBE0SME",
	},
	"76": {
		"4/4": "CQACAgIAAxkBAAEB0vZiLlwZYJQZYNbnffgZAwzVlDTcmQACfRQAAo5qeEmDtzsmt4MV9SME",
	},
	"77": {
		"4/4": "CQACAgIAAxkBAAEB0vhiLlwu_yx0lv8IQv3QnB0KnRnwZAACfhQAAo5qeEnKObBNg1hMdCME",
	},
	"78": {
		"4/4": "CQACAgIAAxkBAAEB0vpiLlxCSq8kNs_-SnlcJacfnYcQRwACfxQAAo5qeEkSnjukkMLR6CME",
	},
	"79": {
		"4/4": "CQACAgIAAxkBAAEB0vxiLlxW6S-6_s8-tzn0Iemon54HugACgBQAAo5qeEl5TQ-LhAEUryME",
	},
	"80": {
		"4/4": "CQACAgIAAxkBAAEB0v5iLlxqAmniShk1l5xM_u7Co8KBdwACgRQAAo5qeEngGS3aFLaAnyME",
	},
	"81": {
		"4/4": "CQACAgIAAxkBAAEB0wABYi5cf_5OamrenlN0UbXnTa2yoQwAAoIUAAKOanhJq8-IUYCI_McjBA",
	},
	"82": {
		"4/4": "CQACAgIAAxkBAAEB0wJiLlyUYNuuLzmpVVF8p-lDWt6JqwACgxQAAo5qeEm-q7ryy2ZX1SME",
	},
	"83": {
		"4/4": "CQACAgIAAxkBAAEB0wRiLlypFl81PzQbJuVmbVu7mJVrPQAChBQAAo5qeEnnMBOZaQj-uyME",
	},
	"84": {
		"4/4": "CQACAgIAAxkBAAEB0wZiLly9ooKkGOWJliYG_ochOTqvugAChRQAAo5qeEkUatUsu7Ly2iME",
	},
	"85": {
		"4/4": "CQACAgIAAxkBAAEB0whiLlzR-ZOu7RJ9h5dgRKAfNN6w8QAChhQAAo5qeEm3TFe6eZaZ9yME",
	},
	"86": {
		"4/4": "CQACAgIAAxkBAAEB0wpiLlzlkbZyAqyLPzK0hrSnMNCBtQAChxQAAo5qeEm6iamFS9tzrSME",
	},
	"87": {
		"4/4": "CQACAgIAAxkBAAEB0wxiLlz33Rx1RIFB32NbxLzchhuvywACiBQAAo5qeElnzzckKr0KTyME",
	},
	"88": {
		"4/4": "CQACAgIAAxkBAAEB0w5iLl0KbMJodU4Cy5AYxYg0teACEgACiRQAAo5qeEkw3I8Pe05vniME",
	},
	"89": {
		"4/4": "CQACAgIAAxkBAAEB0xBiLl0d2zsnnnkSEYJ4FX8e4Vgi7AACihQAAo5qeEmx87-DNjT__yME",
	},
	"90": {
		//"4/4": "CQACAgIAAxkBAALNR2IyMPjZlmeWuxk8Mp4vRsKSg06hAALfEgACvF15SUJM8UiVZVUdIwQ",
		"4/4": "CQACAgIAAxkBAAEB0xJiLl0w-fi8Nif_NSyQPeXtfgxzbQACixQAAo5qeEnz9oNzsIGYSiME",
	},
	"91": {
		"4/4": "CQACAgIAAxkBAAEB0xRiLl1BgEeXHUN2snnhNoAOO-SozQACjBQAAo5qeEl0-NpT_USvTyME",
	},
	"92": {
		"4/4": "CQACAgIAAxkBAAEB0xZiLl1TouWS41Hhgp1joKZ371HfIwACjRQAAo5qeElu4KAZWbyI_yME",
	},
	"93": {
		"4/4": "CQACAgIAAxkBAAEB0xhiLl1laNN7nFRSuOBFrm-6P9mHPQACjhQAAo5qeEnPU5UagFOc1CME",
	},
	"94": {
		"4/4": "CQACAgIAAxkBAAEB0xpiLl134DJfeXrgjqLTLbYYNrn0JAACjxQAAo5qeEkQ0k4Ft6vBMCME",
	},
	"95": {
		"4/4": "CQACAgIAAxkBAAEB0xxiLl2ILKgRrKK2QYpTxU81biAYvwACkBQAAo5qeEnLFcNurYyicSME",
	},
	"96": {
		"4/4": "CQACAgIAAxkBAAEB0x5iLl2X1NvICJPfFlbiO78WM5s69gACkRQAAo5qeEkbNhokwbhIaiME",
	},
	"97": {
		"4/4": "CQACAgIAAxkBAAEB0yBiLl2mZKEQu2jtGM_99Rv1JoYE8wACkxQAAo5qeEmc6Hj4ZQ8OhSME",
	},
	"98": {
		"4/4": "CQACAgIAAxkBAAEB0yJiLl21dx0zynCMCoagMKwm1AXkGwAClBQAAo5qeElQHFOzL559HCME",
	},
	"99": {
		"4/4": "CQACAgIAAxkBAAEB0yRiLl3Eb8eMzVdc0D7pJWPwpNiWOQAClRQAAo5qeEmUQmZpj4WKmSME",
	},
	"100": {
		"4/4": "CQACAgIAAxkBAAEB0yZiLl3VBEh_J0szJCcQLnCs-0ZCYwACchoAAsTVcEloAwHZfZcv5SME",
	},
	"101": {
		"4/4": "CQACAgIAAxkBAAEB0yhiLl3qnxWfzdsECSXUWtwW3DzQNgAClhQAAo5qeEmAoZa5XpLqdCME",
	},
	"102": {
		"4/4": "CQACAgIAAxkBAAEB0ypiLl39Df26gJi-qJ15oZJrjWfl1gACmBQAAo5qeEkJCwABzcnVBwUjBA",
	},
	"103": {
		"4/4": "CQACAgIAAxkBAAEB0yxiLl4QCGWuNdeQrwQV54PnnHwZyQACmhQAAo5qeEn_AjBn6MQcPyME",
	},
	"104": {
		"4/4": "CQACAgIAAxkBAAEB0y5iLl4iEM49Cr2z6jKdaqqqgwABE2MAAp0UAAKOanhJx-zNQ1HQHPcjBA",
	},
	"105": {
		"4/4": "CQACAgIAAxkBAAEB0zBiLl40emnQJynl38vw6ViMktEECQACnhQAAo5qeEl6vf-N-6BxxSME",
	},
	"106": {
		"4/4": "CQACAgIAAxkBAAEB0zJiLl5GPf2FoWF65vbk4CFUuXdP3QACnxQAAo5qeEnk_5ilkaBOGCME",
	},
	"107": {
		"4/4": "CQACAgIAAxkBAAEB0zRiLl5XHqp8CntgRtlrNg6FeymvTgACoBQAAo5qeEkFUMx9GHoWKCME",
	},
	"108": {
		"4/4": "CQACAgIAAxkBAAEB0zZiLl5pVip1K6C5yCGF2KWOHomwpgACoRQAAo5qeElfp28Y0MkRbCME",
	},
	"109": {
		"4/4": "CQACAgIAAxkBAAEB0zhiLl55X5GzKiedQH_b_f75duVJTAACohQAAo5qeEn8N4fH2cnMHCME",
	},
	"110": {
		"4/4": "CQACAgIAAxkBAAEB0zpiLl6a6pQrC0Hbj67Q2KtzuktsiAACoxQAAo5qeEnGwZXVJ_3suSME",
	},
	"111": {
		"4/4": "CQACAgIAAxkBAAEB0zxiLl64O_Hcak6VrEMk7EHJAzlNmAACpBQAAo5qeEkFVrOtF_qCUSME",
	},
	"112": {
		"4/4": "CQACAgIAAxkBAAEB0z5iLl7VVM7plh7CO978DcGkGA2biwACpxQAAo5qeEkWSD-bHRPpICME",
	},
	"113": {
		"4/4": "CQACAgIAAxkBAAEB00BiLl7xlbIhahj92Y5T5qmsa-BzjAACqBQAAo5qeElaD7l_bfRLqSME",
	},
	"114": {
		"4/4": "CQACAgIAAxkBAAEB00JiLl8NGViohscC3yP2UT2zn5UYjwACqRQAAo5qeEnCp30zkJxKCiME",
	},
	"115": {
		"4/4": "CQACAgIAAxkBAAEB00RiLl8lcYUCzEPj_U8oFVsva_zu7gACqhQAAo5qeEk6RXAS8_zqjCME",
	},
	"116": {
		"4/4": "CQACAgIAAxkBAAEB00ZiLl88gTrcxsOxymvSRud5vQW9UQACqxQAAo5qeElJm1JIQSI4_yME",
	},
	"117": {
		"4/4": "CQACAgIAAxkBAAEB00hiLl9VKM57F6B5NNXleVbpzFNUDQACrRQAAo5qeElZbFpRhIU9JyME",
	},
	"118": {
		"4/4": "CQACAgIAAxkBAAEB00piLl9vaW7Wg9jnUT-mWNi4BsmmpQACrhQAAo5qeEmR1ZFSq97qwSME",
	},
	"119": {
		"4/4": "CQACAgIAAxkBAAEB00xiLl-HLHYCRjBH4jpZHEADfT41XQACrxQAAo5qeEnkhrUlgbvRbSME",
	},
	"120": {
		"4/4": "CQACAgIAAxkBAAEB005iLl-hxXHelMthfjFm4snR5-WFQAACsBQAAo5qeEn4CztpHlTD7SME",
	},
	"121": {
		"4/4": "CQACAgIAAxkBAAEB01BiLl-6ZzPIxhAo5hkafUQGaVjTTgACsRQAAo5qeEm-ZeiT3SMD2yME",
	},
	"122": {
		"4/4": "CQACAgIAAxkBAAEB01JiLl_TIh9XgFBBMWwkB_shCYhpqgACshQAAo5qeEkbCN4c46T9xSME",
	},
	"123": {
		"4/4": "CQACAgIAAxkBAAEB01RiLl_rbwzpucqJWHQxZnvOycJszQACsxQAAo5qeEnsG7GzK3it9SME",
	},
	"124": {
		"4/4": "CQACAgIAAxkBAAEB01ZiLmACRLZOXTmygeTUE2qagOaK3QACtBQAAo5qeEm20yryqA0IGyME",
	},
	"125": {
		"4/4": "CQACAgIAAxkBAAEB01hiLmAbtD2Cut_YUwHbiNjyY4X8GgACtRQAAo5qeEl7pL7ODZBoKSME",
	},
	"126": {
		"4/4": "CQACAgIAAxkBAAEB01piLmAnBusadzffxGTompVkmsyUhwACthQAAo5qeEna5H2WMKrQCiME",
	},
	"127": {
		"4/4": "CQACAgIAAxkBAAEB01xiLmAwO4c_q0tbl8GK-l7LSt3LUgACtxQAAo5qeEmxrdKo4F-B_SME",
	},
	"128": {
		"4/4": "CQACAgIAAxkBAAEB015iLmA4GqhEA9nAdpschCq8SG7wfQACuBQAAo5qeEkbzGz90wZ1sCME",
	},
	"129": {
		"4/4": "CQACAgIAAxkBAAEB02BiLmA_8PUdOeOvnaR7pDQ-8bYyiAACuRQAAo5qeEkmW9rj77sFYyME",
	},
	"130": {
		"4/4": "CQACAgIAAxkBAAEB02JiLmBJfeBPzZEdURnuSkBgh2ZO1QACuhQAAo5qeEk4--hNOEjyWiME",
	},
	"131": {
		"4/4": "CQACAgIAAxkBAAEB02RiLmBPvAFQJ_Yqz3cg28Y4oq9SPwACuxQAAo5qeEmVc4GWeN2I7CME",
	},
	"132": {
		"4/4": "CQACAgIAAxkBAAEB02ZiLmBSNLrfjiOu6rtusRPWW1v9iQACvBQAAo5qeEnKpeTkdKVhDCME",
	},
	"133": {
		"4/4": "CQACAgIAAxkBAAEB02hiLmBUQLhe7no9gOKvEdwNIJwyrwACvRQAAo5qeEmCWRLlniM5AyME",
	},
	"134": {
		"4/4": "CQACAgIAAxkBAAEB02piLmBXN_6oVRmEKW1OFj7DSRXm-AACvhQAAo5qeEnrGK4TkgP-SSME",
	},
	"135": {
		"4/4": "CQACAgIAAxkBAAEB02xiLmBekPF3MR96Tjvmkhvs1AEr1gACvxQAAo5qeEnXVAJzkEg8USME",
	},
	"136": {
		"4/4": "CQACAgIAAxkBAAEB025iLmBhG0TXxgI11LfnSyVCkwMTfwACwBQAAo5qeEkTJ97_9qK6xSME",
	},
	"137": {
		"4/4": "CQACAgIAAxkBAAEB029iLmBkYib6LpXXtOFcQhPutXwCIgACwRQAAo5qeEluRnD7mQaSOSME",
	},
	"138": {
		"4/4": "CQACAgIAAxkBAAEB03JiLmCB-RO1B1i2no97FIa8YnZLlQACwxQAAo5qeElszM2fL2bhvSME",
	},
	"139": {
		"4/4": "CQACAgIAAxkBAAEB03RiLmCbuCjJAeiT-T2sBlUpB7HnqwACxBQAAo5qeEk7qUu7Ut0UKSME",
	},
	"140": {
		"4/4": "CQACAgIAAxkBAAEB03ZiLmC4SHWu0MnGFGTepskrss1FSAACxhQAAo5qeEkKuT2VkZ66kSME",
	},
	"141": {
		"4/4": "CQACAgIAAxkBAAEB03hiLmDSb9vtD4nbPC6f3xFL6VWsdQACxxQAAo5qeEm3kw-qbtPJ6iME",
	},
	"142": {
		"4/4": "CQACAgIAAxkBAAEB03piLmDr5JM2GhfNaCtrfNnFuWcQwQACyBQAAo5qeEnJqoNdOntu5SME",
	},
	"143": {
		"4/4": "CQACAgIAAxkBAAEB03xiLmEI_AatjGSqIzWJ2bE80mbUEwACyhQAAo5qeEnYiELoDZtTBiME",
	},
	"144": {
		"4/4": "CQACAgIAAxkBAAEB035iLmElauMyvK3QGLJpD93DAZ-d1AACyxQAAo5qeElJVJP_rBV3yyME",
	},
	"145": {
		"4/4": "CQACAgIAAxkBAAEB04BiLmFBDdwPUWbB-1L3zQcODzBaEAACzBQAAo5qeEldOrUgvDVGQCME",
	},
	"146": {
		"4/4": "CQACAgIAAxkBAAEB04JiLmFEfMrdHuqdKV1b1lqnsd3s_QACzRQAAo5qeElo4iS3ZAXQliME",
	},
	"147": {
		"4/4": "CQACAgIAAxkBAAEB04RiLmFhUz4-tHdiOcKjOIWuEO0mBgACzhQAAo5qeEnG8WjcZbAL8SME",
	},
	"148": {
		"4/4": "CQACAgIAAxkBAAEB04ZiLmF-8W4dG2RIoHTGPEBALFWwPwACzxQAAo5qeEkkG3uxwgABHEEjBA",
	},
	"149": {
		"4/4": "CQACAgIAAxkBAAEB04hiLmGYrXWH_wSg0LHOkiHU6ei2bgAC0BQAAo5qeEl2NmLzxsjzeSME",
	},
	"150": {
		"4/4": "CQACAgIAAxkBAAEB04piLmGzd3z7DPiN673vC0TztSQb7wAC0RQAAo5qeEm6W3bRKPjKGiME",
	},
	"151": {
		"4/4": "CQACAgIAAxkBAAEB04xiLmHMKmNzQHsAAdq9aBpOuoiRIh8AAtIUAAKOanhJaGECLpriPuIjBA",
	},
	"152": {
		"4/4": "CQACAgIAAxkBAAEB045iLmHjq2lCsh2kx5Eqzzi0NvvO2gAC0xQAAo5qeEkCegLjzvGqbyME",
	},
	"153": {
		"4/4": "CQACAgIAAxkBAAEB05BiLmIBaDJCb2Ws1YXMMLSTIDv7vQAC1BQAAo5qeElpw-VELKlFviME",
	},
	"154": {
		"4/4": "CQACAgIAAxkBAAEB05JiLmIdQFGysvM8s06urBZP9fTY0AAC1RQAAo5qeEmU0BMYjnraTSME",
	},
	"155": {
		"4/4": "CQACAgIAAxkBAAEB05RiLmI1Vu2wswdrAAH8jtCMCxmqzD0AAtYUAAKOanhJmjrEEaatup8jBA",
	},
	"156": {
		"4/4": "CQACAgIAAxkBAAEB05ZiLmJLJN834JLDMmnTl7noD4ZTiQAC1xQAAo5qeEmz8VTqHX1SryME",
	},
	"157": {
		"4/4": "CQACAgIAAxkBAAEB05hiLmJiFhNHVyNJoJtyDj5ymSEmxgAC2RQAAo5qeEnYfiq9O95JgCME",
	},
	"158": {
		"4/4": "CQACAgIAAxkBAAEB05piLmJ6TCBvRFNLV7-DU_QsIMUchgAC2hQAAo5qeEnOfJ8huhMB9iME",
	},
	"159": {
		"4/4": "CQACAgIAAxkBAAEB05xiLmKSJ_KP5MhDk6lu64F9c-sCHQAC2xQAAo5qeEnXMSyYQyWkEyME",
	},
	"160": {
		"4/4": "CQACAgIAAxkBAAEB055iLmKsaTqCmgjvVIQI4nNWyUxDWQAC3BQAAo5qeEk0LhHh_0DOlCME",
	},
}

func GetMetronomeTrackFileID(bpm, time string) string {
	bpmObj, ok := metronomeTracks[bpm]
	if !ok {
		// todo
		bpmObj = metronomeTracks["60"]
	}

	track, ok := bpmObj[time]
	if !ok {
		track = bpmObj["4/4"]
	}

	return track
}
