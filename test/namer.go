/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

////
// This awesome code is from the heart and mind of @kris-nova
////

package test

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Iterations of recursive attempted GetClusterName() calls
// The higher this value, the more attempts to get a name
const RecursiveGetClusterNameTimeout = 10

// We define a well know public resolvable IPv4 address here
// we use this in the failFast() function to perform a simple
// connectivity test before entertaining anything AWS API related
// In this case we attempt to open up a basic UDP socket on a
// well known public DNS server. Google's public NS.
const PublicResolvableIpv4 = "8.8.8.8:53"

// For the failFast() functon to work, we must define a timeout duration
// Here is where we set the acceptable amount of time (in milliseconds)
// that we will attempt to wait for a TCP response from PublicResolvableIpv4
// Remember the function is called failFast() so we need to keep this relatively
// short!
const PublicResolvableIpv4FailFastTimeoutDuration = time.Millisecond * 750

// Pass a domain name - get a unique cluster ID
// This is the core of Kops's naming generator
//
// All dependencies are self contained, and you can
// rest assured that your cluster doesn't exist if
// you get a name
func GetRandomClusterName(domain string) (string, error) {
	if err := failFast(); err != nil {
		return "", err
	}
	return getRandomClusterNameRecursive(domain, 0)
}

func getRandomClusterNameRecursive(domain string, recursion int) (string, error) {
	if recursion == RecursiveGetClusterNameTimeout {
		return "", fmt.Errorf("Unable to get unique name after max recursive attempts. See logs.")
	}
	rand.Seed(time.Now().UTC().UnixNano())
	names := strings.Split(MasterNamesList, "\n")
	l := len(names)
	r1 := random(0, l)
	r2 := random(0, l)
	l1 := strings.Split(names[r1], " ")
	l2 := strings.Split(names[r2], " ")

	// This is a sanity check to make sure our list looks okay
	// We have an easy opportunity to try again - so lets do it
	if len(l1) < 1 || len(l2) < 2 {
		fmt.Println(l1, l2)
		log.Printf("Major failure in cluster naming list! Must contain 2 values!")
		return getRandomClusterNameRecursive(domain, recursion+1)
	}

	n1 := l1[0] //Random first name
	n2 := l2[1] // Random last name

	// The magic formula for our cluster names
	// Todo remove punctuation here
	potentialName := fmt.Sprintf("%s-%s.%s", strings.ToLower(n1), strings.ToLower(n2), domain)

	// A lovely example of recursion
	name, err := GetFQN(potentialName)
	if err != nil {
		log.Printf("Major failure in namer duplicate detection: %v", err)
		return getRandomClusterNameRecursive(domain, recursion+1)
	}
	// If we have a name - it already exists, lets try again
	if name != "" {
		return getRandomClusterNameRecursive(domain, recursion+1)
	}
	return potentialName, nil
}

// If we fail, we want to fail fast.. so lets just check BASIC internet connectivity before even
// considering talking to the AWS api..
// Lets try to write an empty to message to a DNS socket and cause some trouble..
func failFast() error {
	conn, err := net.DialTimeout("tcp", PublicResolvableIpv4, PublicResolvableIpv4FailFastTimeoutDuration)
	if err != nil {
		return fmt.Errorf("Unable to confirm connectivity to defined well-known public resolvable server: %v", err)
	}
	bytes := []byte("")
	_, err = conn.Write(bytes)
	if err != nil {
		return fmt.Errorf("Unable to confirm connectivity to defined well-known public resolvable server: %v", err)
	}
	return nil

}

// Generate a random index
func random(min, max int) int {
	return rand.Intn(max-min) + min
}


// A fancy concurrent function for checking all AWS regions for a cluster
func GetFQN(name string) (string, error) {
	ch := make(chan string)
	for region, _ := range AwsRegions {
		go concurrentNameSearch(name, region, ch)
	}
	var names []string
	for i := 0; i < len(AwsRegions); i++ {
		b := <-ch
		if b == "" {
			continue
		}
		names = append(names, b)
	}
	if len(names) == 1 {
		return names[0], nil
	} else if len(names) > 1 {
		return "", fmt.Errorf("More than 1 matching name found: %s", strings.Join(names, ", "))
	}
	return "", nil

}


func concurrentNameSearch(name, region string, ch chan string) {
	aws, err := NewAws(region)
	if err != nil {
		log.Printf("Unable to check region for cluster: %v", err)
		ch <- ""
		return
	}
	result, err := aws.S3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		// Squelching here to quiet logs.. We see this a lot
		log.Printf("Unable to list buckets: %v", err)
		ch <- ""
		return
	}
	for _, bucket := range result.Buckets {
		// Straight up ignore non-kops buckets - we want to go fast!
		if !strings.Contains(*bucket.Name, "kops-test-") {
			continue
		}
		delim := "/"
		request := &s3.ListObjectsInput{Bucket: bucket.Name, Delimiter: &delim}
		ls, err := aws.S3.ListObjects(request)
		if err != nil {
			continue
			// So unsure if this should break or not - for now - lets take the safer route
			// and let the algorithm finish..
			//log.Printf("Unable to list objects from bucket %s: %v", *bucket.Name, err)
			//ch <- ""
			//return
		}
		if len(ls.CommonPrefixes) < 1 {
			continue
		}
		for _, potentialCluster := range ls.CommonPrefixes {
			if strings.Contains(*potentialCluster.Prefix, name) {
				// We found the cluster
				// Remove the trailing / in the cluster
				ch <- strings.Replace(*potentialCluster.Prefix, "/", "", 1)
				return
			}
		}
	}
	ch <- ""
}

// TODO: pull these from aws api
var AwsRegions = map[string][]string{
	"us-west-2":      {"us-west-2a", "us-west-2b", "us-west-2c"},
	"us-west-1":      {"us-west-1b", "us-west-1c"},
	"us-east-1":      {"us-east-1a", "us-east-1c", "us-east-1d"}, // "us-east-1e" Removing for now.  Etcd needs an even number.make
	"us-east-2":      {"us-east-2a", "us-east-2b", "us-east-2c"},
	"eu-west-1":      {"us-west-1b", "us-west-1c"},
	"sa-east-1":      {"sa-east-1a", "sa-east-1c"}, // sa-east-1b does not support Nat Gateways
	"ap-southeast-2": {"ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"},
}

type Aws struct {
	Session *session.Session
	S3      *s3.S3
	EC2     *ec2.EC2
}

// Create new AWS struct. Useful for connection to the API
// This function will look for aws credentials in the correct AWS credential path ~/.aws/credentials
//
// Each AWS is married to a single region
func NewAws(region string) (*Aws, error) {
	a := &Aws{}
	s, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Unable to create new AWS session: %v", err)
	}
	a.Session = s

	// S3
	s3 := s3.New(s, &aws.Config{Region: aws.String(region)})
	a.S3 = s3

	// EC2
	ec2 := ec2.New(s, &aws.Config{Region: aws.String(region)})
	a.EC2 = ec2

	return a, nil
}

//
// Here are the master names to chose from for naming Kubernetes
// clusters.
//
// You are WELCOME to add names to the list but the names MUST meet the following criteria:
//
// Newline delimited name pairs separated by a single space
// No hyphenated names
// No 3 part names
// No 1 part names
// Caps don't matter
//
const MasterNamesList = `
Hippoi Kabeirikoi
The Khalkotauroi
Kourai Khryseai
Golden Dog
Scythian Dracaena
Giantomachian Dragon
Nemean Lion
Satyros Aithiopikos
Satyros Argios
Satyros Lemnios
Corinthian Lamia
Ceryneian Hind
Ethiopian Cetus
Trojan Cetus
Elaphoi Khrysokeroi
Aberforth Dumbledore
Alastor Moody
Albus Dumbledore
Albus Potter
Alecto Carrow
Alice Longbottom
Alicia Spinnet
Amelia Bones
Amos Diggory
Amycus Carrow
Andromeda Tonks
Angelina Johnson
Anthony Goldstein
Antioch Peverell
Antonin Dolohov
Arabella Figg
Argus Filch
Ariana Dumbledore
Arthur Weasley
Augusta Longbottom
Augustus Rookwood
Aurora Sinistra
Bathilda Bagshot
Bathsheba Babbling
Bellatrix Lestrange
Bertha Jorkins
Blaise Zabini
Cadmus Peverell
Cedric Diggory
Charity Burbage
Charlie Weasley
Colin Creevey
Cormac McLaggen
Cornelius Fudge
Crookshanks
Cuthbert Binns
Dean Thomas
Dedalus Diggle
Dennis Creevey
Dilys Derwent
Dirk Cresswell
Dolores Umbridge
Draco Malfoy
Dudley Dursley
Elphias Doge
Emmeline Vance
Ernie Macmillan
Fenrir Greyback
Filius Flitwick
Fleur Delacour
Frank Bryce
Frank Longbottom
Fred Weasley
Gabrielle Delacour
Garrick Ollivander
Gellert Grindelwald
George Weasley
Gilderoy Lockhart
Ginevra Weasley
Godric Gryffindor
Hannah Abbott
Harry Potter
Helena Ravenclaw
Helga Hufflepuff
Hepzibah Smith
Hermione Granger
Horace Slughorn
Ignotus Peverell
Igor Karkaroff
Irma Pince
James Potter
Siruius Potter
Justin Fletchley
Katie Bell
Kingsley Shacklebolt
Lavender Brown
Lee Jordan
Lily Potter
Lily Potter
Lucius Malfoy
Ludo Bagman
Luna Lovegood
Mafalda Hopkirk
Madam Malkin
Marcus Flint
Marietta Edgecombe
Marjorie Dursley
Marvolo Gaunt
Mary Cattermole
Mary Riddle
Merope Gaunt
Michael Corner
Millicent Bulstrode
Minerva McGonagall
Molly Weasley
Morfin Gaunt
Mundungus Fletcher
Narcissa Malfoy
Neville Longbottom
Newt Scamander
Nicolas Flamel
Nymphadora Tonks
Oliver Wood
Olympe Maxime
Padma Patil
Pansy Parkinson
Parvati Patil
Penelope Clearwater
Percy Weasley
Peter Pettigrew
Petunia Dursley
Phineas Nigellus
Pius Thicknesse
Pomona Sprout
Poppy Pomfrey
Quirinus Quirrell
Rabastan Lestrange
Reginald Cattermole
Remus Lupin
Rita Skeeter
Rodolphus Lestrange
Romilda Vane
Ronald Weasley
Madam Rosmerta
Rowena Ravenclaw
Rubeus Hagrid
Rufus Scrimgeour
Salazar Slytherin
Seamus Finnigan
Septima Vector
Sirius Black
Stan Shunpike
Sturgis Podmore
Susan Bones
Ted Tonks
Teddy Lupin
Terry Boot
Theodore Nott
Tom Riddle
Vernon Dursley
Viktor Krum
Lord Voldemort
Walden Macnair
William Weasley
Xenophilius Lovegood
Zacharias Smith
ANAKIN SKYWALKER
DARTH VADER
LUKE SKYWALKER
OBI WAN
CHEWBACCA SOLO
LEIA SKYWALKER
PRINCESS PADME
QUIGON JINN
YODA YODA
JARJAR BINKS
LANDO calrissian
CAPTAIN PANAKA
GENERAL GRIEVOUS
BOBA FETT
JANGO FETT
ADMIRAL ACKBAR
COUNT DOOKU
DARTH MAUL
`
