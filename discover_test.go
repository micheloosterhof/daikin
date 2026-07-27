// ABOUTME: Tests for interpreting UDP discovery responses as Device
// ABOUTME: descriptions with decoded names.
package daikin

import "testing"

const officeBasicInfo = "ret=OK,type=aircon,reg=th,dst=0,ver=1_16_0,rev=182179A,pow=1,err=0,location=0,name=%4f%66%66%69%63%65,icon=5,method=home only,lpw_flag=0,adp_kind=3,pv=2,cpv=2,cpv_minor=00,led=1,en_setzone=1,mac=A841F4D64496,ssid=DaikinAP02770,adp_mode=ap_run,en_hol=0,enlver=1.00,grp_name=,en_grp=0,en_secure=1"

func TestDeviceFromBasicInfo(t *testing.T) {
	dev, err := DeviceFromBasicInfo("192.168.179.237", officeBasicInfo)
	if err != nil {
		t.Fatalf("DeviceFromBasicInfo returned error: %v", err)
	}
	want := Device{IP: "192.168.179.237", Name: "Office", MAC: "A841F4D64496", PowerOn: true}
	if dev != want {
		t.Errorf("DeviceFromBasicInfo = %+v, want %+v", dev, want)
	}
}

func TestDeviceFromBasicInfoBadPayload(t *testing.T) {
	if _, err := DeviceFromBasicInfo("10.0.0.1", "ret=PARAM NG"); err == nil {
		t.Fatal("expected error for non-OK payload")
	}
}
