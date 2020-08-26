package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSMediaLiveInputSecurityGroup_basic(t *testing.T) {
	resourceName := "aws_media_live_input_security_group.test"
	testIP, _ := acctest.RandIpAddress("192.168.0.0/16")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaLive(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaLiveInputSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaLiveInputSecurityGroupConfig(testIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaLiveInputSecurityGroupExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "medialive", regexp.MustCompile(`inputSecurityGroup:[0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv4_whitelist.0", fmt.Sprintf("%v/32", testIP)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMediaLiveInputSecurityGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).medialiveconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_live_input_security_group" {
			continue
		}

		input := &medialive.DescribeInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeInputSecurityGroup(input)
		if err == nil {
			return fmt.Errorf("MediaLive Input Security Group (%s) not deleted", rs.Primary.ID)
		}

		if !isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaLiveInputSecurityGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).medialiveconn

		input := &medialive.DescribeInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeInputSecurityGroup(input)

		return err
	}
}

func testAccMediaLiveInputSecurityGroupConfig(ip string) string {
	return fmt.Sprintf(`
resource "aws_media_live_input_security_group" "test" {
  ipv4_whitelist = ["%v/32"]
}
`, ip)
}
