package main

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

var Account_id = ""
var Repos_to_apply_policy = []string{}
var Policy_filename string = "ecr-policy.json"

func main() {
	create_lifecycle_policy()
	// get_repositories()
}

func createsession() aws.Config {

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return cfg
}

func get_repositories() []string {

	cfg := createsession()
	ecr_client := ecr.NewFromConfig(cfg)
	ecr_describe_repos, err := ecr_client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})

	if err != nil {
		log.Panic(err)
	}

	for _, repo_list := range ecr_describe_repos.Repositories {
		// fmt.Println(aws.ToString(repo_list.RepositoryName)) //prints all the repos found
		if strings.HasPrefix(aws.ToString(repo_list.RepositoryName), "istio") {
			continue
		} else {
			Repos_to_apply_policy = append(Repos_to_apply_policy, aws.ToString(repo_list.RepositoryName))
		}
	}

	// fmt.Println(Repos_to_apply_policy) //prints all the repos to apply the policy on

	return Repos_to_apply_policy

}

func create_lifecycle_policy() {

	//open the file
	openfile, err := os.OpenFile(Policy_filename, os.O_RDONLY, 0600) //os.OpenFile(name of file, which mode is the file opened on, permissions)
	if err != nil {
		panic(err)
	}

	defer openfile.Close()

	//read the file contents
	content, err := io.ReadAll(openfile)
	if err != nil {
		panic(err)
	}

	cfg := createsession()
	Repos_to_apply_policy := get_repositories()

	ecr_client := ecr.NewFromConfig(cfg)

	for i := 0; i < len(Repos_to_apply_policy); i++ {

		repo_to_add := Repos_to_apply_policy[i]

		ecr_create_lifecycle_output, err := ecr_client.PutLifecyclePolicy(context.TODO(), &ecr.PutLifecyclePolicyInput{
			// The JSON repository policy text to apply to the repository.
			// LifecyclePolicyText is a required field
			LifecyclePolicyText: aws.String(string(content)),
			RegistryId:          aws.String(Account_id),  //account ID
			RepositoryName:      aws.String(repo_to_add), // name of the repository
		})

		if err != nil {
			log.Panic(err)
		}

		log.Printf("Lifecycle policy successfully applied to %v repo", aws.ToString(ecr_create_lifecycle_output.RepositoryName))
	}
}
