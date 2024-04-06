package sample

import (
	"fmt"
	"oos-go-sdk/oos"
)

/*
<LifecycleConfiguration>
	<Rule>
		<Prefix>logs</Prefix>
		<Status>Enabled</Status>
		<Transition>
			<Days>10</Days>
			<StorageClass>STANDARD_IA</StorageClass>
		</Transition>
		<Expiration>
			<Days>20</Days>
		</Expiration>
	</Rule>
</LifecycleConfiguration>
*/
// BucketLifecycleSample shows how to set, get and delete bucket's lifecycle.
func BucketLifecycleSample() {
	// New client
	client := NewClient()

	// Case 1: Set the lifecycle. The rule ID is id1 and the applied objects' prefix is one and expired time is 12/18/2022
	var rule1 = oos.BuildLifecycleExpirRuleByDate("id1", "prefix-test/", false, 2022, 12, 18)
	var rules = []oos.LifecycleRule{rule1}
	err := client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Set the lifecycle, The rule ID is id2 and the applied objects' prefix is two and the expired time is three days after the object created.
	var rule2 = oos.BuildLifecycleExpirRuleByDays("id2", "two", true, 3)
	rules = []oos.LifecycleRule{rule2}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Create two rules in the bucket for different objects. The rule with the same ID will be overwritten.
	var rule3 = oos.BuildLifecycleExpirRuleByDays("id1	", "two", true, 365)
	var rule4 = oos.BuildLifecycleExpirRuleByDate("id2", "one", true, 2022, 12, 13)
	rules = []oos.LifecycleRule{rule3, rule4}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	//case 4 set the liefcycle  one transtion rule
	var rule5 = oos.BuildLifecycleTransitionRuleByDate("id1", "prefix-test/", false, 2022, 12, 18, string(oos.StorageClassStandardIA))
	rules = []oos.LifecycleRule{rule5}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	//case 5
	var rule6 = oos.BuildLifecycleTransitionRuleByDays("id2", "two", true, 30, string(oos.StorageClassStandardIA))
	rules = []oos.LifecycleRule{rule6}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	//case 6 Transition and Expir
	var rule7 = oos.BuildLifecycleExpirRuleByDays("id1", "two", true, 365)
	var rule8 = oos.BuildLifecycleExpirRuleByDate("id2", "one", true, 2022, 12, 13)
	var rule9 = oos.BuildLifecycleTransitionRuleByDays("id3", "two", true, 30, string(oos.StorageClassStandardIA))
	var rule10 = oos.BuildLifecycleTransitionRuleByDate("id4", "prefix-test/", false, 2022, 12, 18, string(oos.StorageClassStandardIA))
	rules = []oos.LifecycleRule{rule7, rule8, rule9, rule10}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	gbl, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Lifecycle:", gbl.Rules)

	// // Delete bucket's Lifecycle
	err = client.DeleteBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketLifecycleSample completed")
}
